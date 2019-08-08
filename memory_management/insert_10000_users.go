package main

import (
	"github.com/ZhengHe-MD/learn-bolt/api/lib"
	"github.com/boltdb/bolt"
	"log"
)

func main() {
	db, _ := bolt.Open("mm1.db", 0600, nil)
	defer db.Close()

	store := lib.NewStore(db)

	_ = store.CleanupBuckets()
	_ = store.EnsureBuckets()

	n := 10000
	if err := store.GenerateFakeUserDataConcurrently(n, 16); err != nil {
		log.Fatal(err)
	}
}
