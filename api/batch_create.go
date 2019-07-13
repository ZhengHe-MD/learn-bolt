package main

import (
	"fmt"
	"github.com/ZhengHe-MD/learn-bolt/api/lib"
	"github.com/boltdb/bolt"
	"log"
	"time"
)

func estimateDuration(d *lib.Dao, title string, f func(d *lib.Dao) error) (err error) {
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
	db, err := bolt.Open("1.DB", 0600, &bolt.Options{
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

	d := lib.NewDao(db)

	_ = d.CleanupBuckets()

	n := 1000

	_ = estimateDuration(d, "insert data one by one", func(d *lib.Dao) error {
		return d.GenerateFakeData(n)
	})

	for ng := 1; ng <= 100; ng++ {
		_ = estimateDuration(
			d,
			fmt.Sprintf("insert data in batch, with %d goroutine", ng),
			func(d *lib.Dao) error {
				return d.GenerateFakeDataConcurrently(n, ng)
			})
	}
}
