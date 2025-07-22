package bboltx

import "go.etcd.io/bbolt"

type BucketPath = []string

func AutoCreate(db *bbolt.DB, paths ...BucketPath) error {
	return db.Update(func(tx *bbolt.Tx) error {
		for _, path := range paths {
			if len(path) == 0 {
				continue
			}
			b, err := tx.CreateBucketIfNotExists([]byte(path[0]))
			if err != nil {
				return err
			}
			for _, key := range path[1:] {
				b, err = b.CreateBucketIfNotExists([]byte(key))
				if err != nil {
					return err
				}
			}
		}
		return nil
	})
}
