package main

import (
	"fmt"
	"github.com/ZhengHe-MD/learn-bolt/api/lib"
	"github.com/boltdb/bolt"
	"log"
	"time"
)

func estimateDuration(d *lib.Store, title string, f func(d *lib.Store) error) (err error) {
	if err = d.EnsureBuckets(); err != nil {
		return
	}

	st := time.Now()
	if err = f(d); err != nil {
		return
	}
	log.Printf("%s takes:%v", title, time.Since(st))

	if err = d.CleanupBuckets(); err != nil {
		return
	}

	return
}

func main() {
	db, err := bolt.Open("2.db", 0600, &bolt.Options{
		// 进程 Open DB 会给 DB 文件加锁，只有一个进
		Timeout:    0,
		NoGrowSync: false,
		// 只读副本使用
		ReadOnly:        false,
		MmapFlags:       0,
		InitialMmapSize: 0,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	store := lib.NewStore(db)

	_ = store.CleanupBuckets()

	n := 1000

	for ng := 1; ng <= 100; ng++ {
		_ = estimateDuration(
			store,
			fmt.Sprintf("insert data in batch, with %d goroutine", ng),
			func(d *lib.Store) error {
				return d.GenerateFakeUserDataConcurrently(n, ng)
			})
	}
}
