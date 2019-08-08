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

	for i := 5000; i < 8000; i++ {
		if err := store.Users.DeleteUserByID(uint64(i)); err != nil {
			log.Fatal(err)
		}
	}
}
