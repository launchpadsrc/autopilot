package cache

import (
	"time"

	"go.etcd.io/bbolt"

	"launchpad.icu/autopilot/parsers"
	"launchpad.icu/autopilot/pkg/bboltx"
)

const (
	keyJobs = "jobs"
)

var buckets = []bboltx.BucketPath{
	{keyJobs, parsers.Dou{}.Host()},
	{keyJobs, parsers.Djinni{}.Host()},
}

var config = &bbolt.Options{
	Timeout: 3 * time.Second,
}

type Cache struct {
	db *bbolt.DB
}

func New(path string) (*Cache, error) {
	db, err := bbolt.Open(path, 0600, config)
	if err != nil {
		return nil, err
	}
	if err := bboltx.AutoCreate(db, buckets...); err != nil {
		return nil, err
	}
	return &Cache{db: db}, nil
}
