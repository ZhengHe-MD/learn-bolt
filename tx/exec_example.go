package main

import (
	"fmt"
	"github.com/boltdb/bolt"
	"log"
)

func main() {
	db, err := bolt.Open("1.db", 0600, nil)
	if err != nil {
		log.Fatal(err)
	}

	err = db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte("b1"))
		if err != nil {
			return err
		}

		return bucket.Put([]byte("k1"), []byte("v1"))
	})

	if err != nil {
		log.Fatal(err)
	}

	err = db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("b1"))
		if bucket != nil {
			v := bucket.Get([]byte("k1"))
			fmt.Printf("%s\n", v)
		}
		return nil
	})
}
