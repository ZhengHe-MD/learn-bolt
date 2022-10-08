# learn-boltdb

在最近的闲暇时间 (2019 年初)，我开始弥补自己数据库知识的盲点。这里也推荐一下自己最喜欢的课程和书籍：

* CMU 15-445/645 Intro to Database Systems
* Designing Data-Intensive Applications (DDIA)

尽管站在巨人的肩膀上，可以体会到数据库的博大精深，但这些努力始终只是在外部远观，并未真正进入其中。

> What I cannot create, I do not understand. --- Richard Feynman

我一直将 Richard Feynman 的这句话记在心中，时刻提醒自己作为软件工程师，不能只满足于看懂，还要能做出来。在动手写一个数据库之前，我想先认真阅读某个数据库的源码，将那些常被提及的概念与实现关联。

在 DDIA 中，Martin Kleppmann 曾多次提及 Bolt，印象颇深的一点是它通过单线程执行读写事务的方式简单粗暴地实现最苛刻的可序列化事务隔离级别。后来的某一天，我心血来潮地读起 Bolt 项目的源码，发现它极简的设计理念正好满足我的学习需求。在阅读过程中，为梳理源码中的概念和逻辑，我开始将自己的理解转化成短文，逐渐形成了本项目。目前，Bolt 项目已经归档，不再更新，这也保证了本文内容不会过时。

## 为什么选择 Bolt

> The original goal of Bolt was to provide a simple pure Go key/value store and to not bloat the code with extraneous features.  --- Ben Johnson

Bolt 可能是最适合 Go 语言工程师阅读的第一个数据库项目，原因在于它功能简单：

* 单机部署：没有像 Raft、Paxos 这样的共识协议
* 无服务端：没有通信协议，直接读写数据库文件
* 存储键值：没有关系代数，没有数据字典，只有键和值
* 无跨语言：仅支持 Go SDK 访问，没有 DSL

实现也简单：

* 数据结构：所有数据存储在同一个树形结构中
* 索引结构：键值数据天然地只有主键索引，使用经典的 B+ 树
* 事务隔离：仅允许多个只读事务和最多一个读写事务同时运行
* 缓存管理：仅管理写缓存，利用 mmap 管理读缓存

## 阅读建议

每篇短文覆盖一个话题，描述对应模块的实现。本系列文章将自底向上地介绍 Bolt，各个模块相对独立，顺序阅读和单篇阅读皆可。

| 主题                              | 源码                                |
| ------------------------------- | --------------------------------- |
| [存储与缓存](./STORAGE_AND_CACHE.md) | page.go, freelist.go, bolt_xxx.go |
| [数据与索引](./DATA_AND_INDEX.md)    | node.go                           |
| [桶](./BUCKET.md)                | bucket.go, cursor.go              |
| [事务](./TX.md)                   | tx.go                             |
| [API] [TODO]                    | db.go                             |

## 名词解释

### 中英对照

为避免专业术语的名称歧义，除已经有稳定中文翻译的词语外，其它词语将保留原始英文形式。下面是一份核心专业词语的中英对照表，供读者参考：

| 中文           | 英文                                   |
| ------------ | ------------------------------------ |
| B+ 树         | B+Tree                               |
| 事务           | transaction/tx                       |
| 读写事务         | read-write transaction/tx            |
| 只读事务         | read-only transaction/tx             |
| 隐式事务         | managed/implicit transaction/tx     |
| 显式事务       | explicit transaction/tx              |
| 提交          | commit |
| 回滚          | rollback |
| 桶            | bucket                               |
| 游标           | cursor                               |
| 键/值/键值对/键值数据 | key/value/key-value(kv) pair/kv data |
| 分裂           | split                                |
| 合并           | merge                                |

### Page

块存储中的数据通常会被分割为等长的数据块，作为读写的最小单元。教科书称之为 frame 或 page frame，这是**物理概念**。当这些数据块被读入内存，在程序中被引用、读写时，教科书称之为 page，那是**逻辑概念**。在 Bolt 中，二者都被称为 page，如:

* 常数 pageSize 中的 page 是物理概念
* page.go 中提到的 page 多是逻辑概念

在阅读源码的过程中要注意区分。

### Bolt 与 Bolt 实例

Bolt 指这个[项目](https://github.com/boltdb/bolt)本身，而 Bolt 实例指利用 api 创建的一个新的数据库，也指文件系统中存储对应数据的二进制文件。

## 参考

* Code
  * [Github: bolt/boltdb](https://github.com/boltdb/bolt)
* Video
  - [CMU 15-445/645 Intro to Database Systems](https://www.youtube.com/playlist?list=PLSE8ODhjZXja3hgmuwhf89qboV1kOxMx7)
* Book
  - [Designing Data-Intensive Applications](https://dataintensive.net/)

转载请注明出处为本项目: https://github.com/ZhengHe-MD/learn-bolt
