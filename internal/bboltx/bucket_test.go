package bboltx_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.etcd.io/bbolt"

	"launchpad.icu/autopilot/internal/bboltx"
)

type testValue struct {
	UID  string `json:"id"`
	Name string `json:"name"`
}

func (v testValue) ID() string {
	return v.UID
}

func openDB(t *testing.T, paths ...bboltx.BucketPath) *bbolt.DB {
	t.Helper()

	dir := t.TempDir()
	path := filepath.Join(dir, "test.db")

	db, err := bbolt.Open(path, 0600, nil)
	require.NoError(t, err)

	err = bboltx.AutoCreate(db, paths...)
	require.NoError(t, err)

	t.Cleanup(func() {
		db.Close()
		os.Remove(path)
	})

	return db
}

func TestPutUniqueAndWalk(t *testing.T) {
	db := openDB(t, bboltx.BucketPath{"users"})
	bucket := bboltx.NewBucket[testValue](db, "users")

	// First insert: all values must be stored.
	vals := []testValue{{"1", "Alice"}, {"2", "Bob"}, {"3", "Carol"}}
	stored, err := bucket.PutUnique(vals)
	require.NoError(t, err)
	assert.Len(t, stored, len(vals))

	// Second insert re‑introduces duplicates and one fresh value.
	more := []testValue{{"2", "Bob"}, {"3", "Carol"}, {"4", "Dave"}}
	stored, err = bucket.PutUnique(more)
	require.NoError(t, err)
	assert.Len(t, stored, 1)
	assert.Equal(t, "4", stored[0].UID)

	// Walk should expose the four unique items now in the DB.
	got := map[string]testValue{}
	err = bucket.Walk(func(v testValue) error {
		got[v.UID] = v
		return nil
	})
	require.NoError(t, err)
	assert.Len(t, got, 4)
}

func TestPutUnique_EmptyID(t *testing.T) {
	db := openDB(t, bboltx.BucketPath{"projects"})
	bucket := bboltx.NewBucket[testValue](db, "projects")
	_, err := bucket.PutUnique([]testValue{{"", "no‑id"}})
	assert.Error(t, err)
}

func TestNestedBucket(t *testing.T) {
	db := openDB(t, bboltx.BucketPath{"root", "sub1", "sub2"})
	nested := bboltx.NewBucket[testValue](db, "root").Bucket("sub1").Bucket("sub2")

	vals := []testValue{{"x1", "nested"}}
	_, err := nested.PutUnique(vals)
	require.NoError(t, err)

	var collected []testValue
	err = nested.Walk(func(v testValue) error {
		collected = append(collected, v)
		return nil
	})
	require.NoError(t, err)
	assert.Equal(t, vals, collected)
}
