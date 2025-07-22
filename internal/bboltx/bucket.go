package bboltx

import (
	"encoding/json"
	"errors"

	"go.etcd.io/bbolt"
)

// BucketValue makes ID field mandatory for bucket values.
type BucketValue interface {
	BoltID() string
}

// Bucket represent a generics-powered bucket for bbolt.
type Bucket[T BucketValue] struct {
	db  *bbolt.DB
	key string
	sub []string
}

// NewBucket creates a new instance of generics-powered bucket.
func NewBucket[T BucketValue](db *bbolt.DB, key string) Bucket[T] {
	return Bucket[T]{
		db:  db,
		key: key,
	}
}

// Bucket returns a wrapper for sub-bucket.
func (b Bucket[T]) Bucket(key string) Bucket[T] {
	return Bucket[T]{db: b.db, key: b.key, sub: append(b.sub, key)}
}

// PutUnique puts values not yet present. Returns the newly stored values.
func (b Bucket[T]) PutUnique(vs []T) (stored []T, _ error) {
	return stored, b.db.Update(func(tx *bbolt.Tx) error {
		bucket := b.bucket(tx)
		for _, v := range vs {
			id := v.BoltID()
			if id == "" {
				return errors.New("bboltx: empty unique id")
			}

			idb := []byte(id)
			if bucket.Get(idb) != nil {
				continue // duplicate
			}

			data, _ := json.Marshal(v)
			if err := bucket.Put(idb, data); err != nil {
				return err
			}

			stored = append(stored, v)
		}
		return nil
	})
}

// Walk iterates over the bucket values.
func (b Bucket[T]) Walk(fn func(T) error) error {
	return b.db.View(func(tx *bbolt.Tx) error {
		return b.bucket(tx).ForEach(func(_, data []byte) error {
			var v T
			if err := json.Unmarshal(data, &v); err != nil {
				return err
			}
			return fn(v)
		})
	})
}

// Count returns the number of items in the bucket.
func (b Bucket[T]) Count() int {
	var count int
	_ = b.db.View(func(tx *bbolt.Tx) error {
		count = b.bucket(tx).Stats().KeyN
		return nil
	})
	return count
}

func (b Bucket[T]) bucket(tx *bbolt.Tx) *bbolt.Bucket {
	bucket := tx.Bucket([]byte(b.key))
	if len(b.sub) > 0 {
		for _, sub := range b.sub {
			bucket = bucket.Bucket([]byte(sub))
		}
	}
	return bucket
}
