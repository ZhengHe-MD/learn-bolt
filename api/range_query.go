package main

import (
	"fmt"
	"github.com/ZhengHe-MD/learn-bolt/api/lib"
	"github.com/boltdb/bolt"
	"log"
	"time"
)

func main() {
	db, err := bolt.Open("3.db", 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	store := lib.NewStore(db)

	_ = store.CleanupBuckets()
	_ = store.EnsureBuckets()

	n := 1000

	if err = store.GenerateFakeEventConcurrently(n, 8); err != nil {
		log.Fatal(err)
	}

	start := time.Now().Unix() - 60*60*24*90
	end := time.Now().Unix() - 60*60*24*30

	events, err := store.Events.GetEventsBetween(start, end)
	if err != nil {
		log.Fatal(err)
	}

	for _, event := range events {
		if event.Time < start || event.Time > end {
			log.Printf("event:%v should not be in events", event)
		}
	}

	fmt.Printf("total:%d", len(events))
}
