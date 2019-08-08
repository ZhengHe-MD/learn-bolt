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
	defer db.Close()

	err = db.Update(func(tx *bolt.Tx) error {
		b1, err := tx.CreateBucketIfNotExists([]byte("b1"))
		if err != nil {
			return err
		}

		_, err = b1.CreateBucketIfNotExists([]byte("b11"))
		if err != nil {
			return err
		}

		_ = b1.Put([]byte("k1"), []byte("v1"))
		_ = b1.Put([]byte("k2"), []byte("v2"))

		return b1.ForEach(func(k, v []byte) error {
			fmt.Println(string(k), string(v))
			return nil
		})
	})

	if err != nil {
		log.Fatal(err)
	}
}
