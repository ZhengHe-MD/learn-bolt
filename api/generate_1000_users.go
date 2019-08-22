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
	defer db.Close()

	store := lib.NewStore(db)
	if err = store.EnsureBuckets(); err != nil {
		log.Fatal(err)
	}

	err = store.GenerateFakeUserDataConcurrently(1000, 20)
	if err != nil {
		log.Fatal(err)
	}
}
