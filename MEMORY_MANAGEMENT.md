# 内存管理

boltDB 运行时可能需要需要：

* 缓冲读数据

* 读写所需临时内存空间

这点我们可以在 boltDB 的 [DB 结构体](https://github.com/boltdb/bolt/blob/master/db.go#L45)中观察得到：

```go
type DB struct {
  // ...
  dataref  []byte                // 缓冲读数据
  data     *[maxMapSize]byte     // 缓冲读数据

  pagePool sync.Pool             // 读写所需临时内存空间
  // ...
}
```

## 缓冲读数据

### mmap

boltDB 使用 mmap 来管理数据的读缓冲。对于 boltDB 来说，数据库文件会被 mmap 映射到内存中的一块区域，boltDB 只管像访问数组一样访问这块区域即可，操作系统会在背后使用 demand paging 的方式将磁盘中的数据按需加载到该内存区域中。当然，使用 mmap 也失去了数据读取的控制权，无法根据数据库的系统知识在数据预取（prefetching）、缓存置换（buffer replacement）方面优化读取性能。

```go
func mmap(db *DB, sz int) error {
  b, err := syscall.Mmap(int(db.file.Fd()), 0, sz, syscall.PROT_READ, syscall.MAP_SHARED|db.MmapFlags)
  if err != nil {
    return err
  }
  
  if err := madvise(b, syscall.MADV_RANDOM); err != nil {
    return fmt.Errorf("madvis: %s", err)
  }
  
  db.dataref = b
  db.data = (*[maxMapSize]byte)(unsafe.Pointer(&b[0]))
  db.datasz = sz
  return nil
}
```

- prot 参数仅使用 syscall.PROT_READ，表示它只对 mmap 后的内存区域执行读操作，不用 mmap 写数据
- flags 参数使用 syscall.MAP_SHARED 表明 mmap 后的内存区域对于所有 boltDB 进程共享。但由于 mmap 的数据是只读的，因此这种共享的作用可能只有节省空间。(参考 mmap)
- madvise 告诉系统这片内存区域的读取模式是随机读取，系统可以据此做相应的优化

mmap 之后，boltDB 就能够通过 db.data 自由地访问任意 page：

```go
// page retrieves a page reference from the mmap based on the current page size.
func (db *DB) page(id pgid) *page {
	pos := id * pgid(db.pageSize)
	return (*page)(unsafe.Pointer(&db.data[pos]))
}
```

在数据存储层，我们见过 db.page 方法：

```go
// meta pages
db.page(0)
db.page(1)
```

## 读写所需临时内存空间

当事务读写数据需要临时内存空间时，就会调用 DB 结构体的 allocate 方法：

```go
func (db *DB) allocate(count int) (*page, error) {
	// Allocate a temporary buffer for the page.
	var buf []byte
	if count == 1 {
		buf = db.pagePool.Get().([]byte)
	} else {
		buf = make([]byte, count*db.pageSize)
	}
	p := (*page)(unsafe.Pointer(&buf[0]))
	p.overflow = uint32(count - 1)

	// Use pages from the freelist if they are available.
	if p.id = db.freelist.allocate(count); p.id != 0 {
		return p, nil
	}

	// Resize mmap() if we're at the end.
	p.id = db.rwtx.meta.pgid
	var minsz = int((p.id+pgid(count))+1) * db.pageSize
	if minsz >= db.datasz {
		if err := db.mmap(minsz); err != nil {
			return nil, fmt.Errorf("mmap allocate error: %s", err)
		}
	}

	// Move the page id high water mark.
	db.rwtx.meta.pgid += pgid(count)

	return p, nil
}
```

事务申请内存空间的过程分为 2 个步骤：

1. 申请内存块：如果事务申请的内存空间为 1 个 page，则可利用 pagePool 来获取，避免频繁地向 heap 申请和释放内存空间；如果事务申请多个 page，则直接从 heap 获取相应的内存块。
2. 标记内存块：分配内存块之后，boltDB 需要能够标记并跟踪这些已经分配的内存块：如果 freelist 中还存在大小符合的连续 的 page id，就可以将这些 id 赋予刚分配的内存空间；否则 boltDB 需要临时扩容，重新评估 db 所需磁盘空间，并重新 mmap 该空间。注意，事务在申请内存空间时，尚未 commit，因此这部分新增的临时空间不会被持久化，只有在 commit 之后，新增的部分才会落盘。

## 参考

##### mmap

- [stackoverflow: is-there-a-difference-between-mmap-map-shared-and-map-private-when-prot-read-is](https://stackoverflow.com/questions/14419940/is-there-a-difference-between-mmap-map-shared-and-map-private-when-prot-read-is)
- [linux: mmap2](http://man7.org/linux/man-pages/man2/mmap.2.html)