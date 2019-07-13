package lib

import (
	"github.com/boltdb/bolt"
	"github.com/brianvoe/gofakeit"
	"log"
	"sync"
	"time"
)

const (
	BucketUsers = "Users"
	BucketEvents = "Events"
)

var Buckets = []string{
	BucketUsers,
	BucketEvents,
}

type Dao struct {
	Users   *UserDao
	Events  *EventDao
	Buckets *BucketDao
}

func NewDao(db *bolt.DB) *Dao {
	return &Dao{
		Users:   NewUserDao(db),
		Events:  NewEventDao(db),
		Buckets: NewBucketDao(db),
	}
}

func (d *Dao) EnsureBuckets() error {
	for _, bucket := range Buckets {
		if err := d.Buckets.CreateBucket([]byte(bucket)); err != nil {
			return err
		}
	}
	return nil
}

func (d *Dao) CleanupBuckets() error {
	for _, bucket := range Buckets {
		if err := d.Buckets.DeleteBucket([]byte(bucket)); err != nil {
			return err
		}
	}
	return nil
}

func newFakeUser() *User {
	return &User{
		Name:   gofakeit.Name(),
		Gender: uint8(gofakeit.Number(0, 1)),
		Age:    uint8(gofakeit.Number(0, 100)),
		Phone:  gofakeit.Phone(),
		Email:  gofakeit.Email(),
		CreatedAt: randInt64Range(
			time.Now().Unix()-60*60*24*365,
			time.Now().Unix(),
		),
	}
}

func newFakeEvent() *Event {
	return &Event{
		Time:   randInt64Range(
			time.Now().Unix()-60*60*24*365,
			time.Now().Unix()),
		Name:   gofakeit.Name(),
		Type:   gofakeit.Uint8(),
		Cancel: gofakeit.Bool(),
	}
}

func (d *Dao) GenerateFakeUserData(n int) error {
	for i := 0; i < n; i++ {
		if err := d.Users.CreateUser(newFakeUser()); err != nil {
			return err
		}
	}
	return nil
}

func (d *Dao) GenerateFakeEventData(n int) error {
	for i := 0; i < n; i++ {
		if err := d.Events.CreateEvent(newFakeEvent()); err != nil {
			return err
		}
	}
	return nil
}

func (d *Dao) generateFakeUserDataInBatch(n int) error {
	for i := 0; i < n; i++ {
		if err := d.Users.CreateUserInBatch(newFakeUser()); err != nil {
			return err
		}
	}
	return nil
}

func (d *Dao) generateFakeEventDataInBatch(n int) error {
	for i := 0; i < n; i++ {
		if err := d.Events.CreateEventInBatch(newFakeEvent()); err != nil {
			return err
		}
	}
	return nil
}

// n should be multiples of ng
func (d *Dao) GenerateFakeDataConcurrently(n int, ng int) error {
	var wg sync.WaitGroup
	wg.Add(ng)
	for i := 0; i < ng; i++ {
		go func() {
			_ = d.generateFakeUserDataInBatch(n/ng)
			wg.Done()
		}()
	}
	wg.Wait()
	return nil
}

func (d *Dao) GenerateFakeEventConcurrently(n int, ng int) error {
	var wg sync.WaitGroup
	wg.Add(ng)
	for i := 0; i < ng; i++ {
		go func() {
			_ = d.generateFakeEventDataInBatch(n/ng)
			wg.Done()
		}()
	}
	wg.Wait()
	return nil
}

func init() {
	log.Println("seed gofakeit")
	gofakeit.Seed(time.Now().UnixNano())
}