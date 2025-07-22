package bboltx_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"go.etcd.io/bbolt"

	"launchpad.icu/autopilot/internal/bboltx"
)

func TestAutoCreate(t *testing.T) {
	db := openDB(t)

	path := bboltx.BucketPath{"a", "b", "c"}
	// First creation should succeed.
	require.NoError(t, bboltx.AutoCreate(db, path))
	// Second creation of the same path should also succeed without error.
	require.NoError(t, bboltx.AutoCreate(db, path))

	// Ensure the hierarchy still exists.
	err := db.View(func(tx *bbolt.Tx) error {
		a := tx.Bucket([]byte("a"))
		b := a.Bucket([]byte("b"))
		c := b.Bucket([]byte("c"))
		require.NotNil(t, a)
		require.NotNil(t, b)
		require.NotNil(t, c)
		return nil
	})
	require.NoError(t, err)
}
