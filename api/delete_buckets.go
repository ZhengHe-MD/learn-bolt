package main

import (
	"github.com/ZhengHe-MD/learn-bolt/api/lib"
	"github.com/boltdb/bolt"
	"log"
)

func main() {
	db, err := bolt.Open("2.db", 0600, nil)
	if err != nil {
		log.Fatal(err)
	}

	store := lib.NewStore(db)

	_ = store.CleanupBuckets()
}
