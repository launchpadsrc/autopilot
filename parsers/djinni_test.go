package parsers

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDjinniParseFeed(t *testing.T) {
	entries, err := NewDjinni().ParseFeed()
	require.NoError(t, err)
	require.NotEmpty(t, entries)
	require.NotEmpty(t, entries[0].ID)
	require.NotEmpty(t, entries[0].Title)
	require.NotEmpty(t, entries[0].Link)
	require.NotEmpty(t, entries[0].Published)
	require.NotEmpty(t, entries[0].Description)
}
