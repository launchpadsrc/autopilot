package cache

import (
	"launchpad.icu/autopilot/parsers"
	"launchpad.icu/autopilot/pkg/bboltx"
)

func (c Cache) StoreFeed(key string, entries []parsers.FeedEntry) ([]parsers.FeedEntry, error) {
	return bboltx.
		NewBucket[parsers.FeedEntry](c.db, keyJobs).
		Bucket(key). // navigate to sub-bucket
		PutUnique(entries)
}
