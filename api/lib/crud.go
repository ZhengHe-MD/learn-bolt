package lib

import (
	"encoding/json"
	"github.com/boltdb/bolt"
)

type User struct {
	ID        uint64
	Name      string
	Gender    uint8
	Age       uint8
	Phone     string
	Email     string
	CreatedAt int64
}

type UserDao struct {
	DB *bolt.DB
}

func NewUserDao(db *bolt.DB) *UserDao {
	return &UserDao{db}
}

func (d *UserDao) genCreateUserFunc(user *User) func(*bolt.Tx) error {
	return func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(BucketUsers))

		id, err := b.NextSequence()
		if err != nil {
			return err
		}
		user.ID = id

		buf, err := json.Marshal(user)
		if err != nil {
			return err
		}

		return b.Put(uint64tob(user.ID), buf)
	}
}

func (d *UserDao) CreateUser(user *User) error {
	return d.DB.Update(d.genCreateUserFunc(user))
}

func (d *UserDao) CreateUserInBatch(user *User) error {
	return d.DB.Batch(d.genCreateUserFunc(user))
}

func (d *UserDao) GetUserByID(id uint64) (user *User, err error) {
	user = &User{}
	err = d.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(BucketUsers))

		buf := b.Get(uint64tob(id))

		bufCopy := make([]byte, len(buf))
		copy(bufCopy, buf)

		err := json.Unmarshal(bufCopy, user)

		if err != nil {
			return err
		}
		return nil
	})
	return
}

func (d *UserDao) GetUsers() (users []*User, err error) {
	err = d.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(BucketUsers))

		return b.ForEach(func(k, v []byte) error {
			var user User
			if err := json.Unmarshal(v, &user); err != nil {
				return err
			}
			users = append(users, &user)
			return nil
		})
	})

	return
}

func (d *UserDao) PutUser(user *User) error {
	return d.DB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(BucketUsers))

		buf, err := json.Marshal(user)
		if err != nil {
			return err
		}

		return b.Put(uint64tob(user.ID), buf)
	})
}

func (d *UserDao) DeleteUserByID(id uint64) error {
	return d.DB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(BucketUsers))
		return b.Delete(uint64tob(id))
	})
}

type BucketDao struct {
	DB *bolt.DB
}

func NewBucketDao(db *bolt.DB) *BucketDao {
	return &BucketDao{db}
}

func (d *BucketDao) CreateBucket(name []byte) error {
	return d.DB.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(name)
		return err
	})
}

func (d *BucketDao) DeleteBucket(name []byte) error {
	return d.DB.Update(func(tx *bolt.Tx) error {
		return tx.DeleteBucket(name)
	})
}