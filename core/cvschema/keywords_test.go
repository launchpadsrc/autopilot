package cvschema

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func newResume(summary string) Resume {
	return Resume{
		Basics: Basics{
			Summary: summary,
		},
		Skills: []Skill{
			{Name: "Golang", Keywords: []string{"concurrency", "sql"}},
		},
	}
}

func TestKeywords_RemovesShortTokensAndStopwords(t *testing.T) {
	r := newResume("Go developer building REST APIs and microservices with Go.")

	got := r.Keywords()

	want := []string{
		"developer", "building", "rest", "apis", "microservices",
		"golang", "concurrency", "sql",
	}

	require.Equal(t, want, got,
		"unexpected token slice\nwant: %#v\n got: %#v", want, got)
}

func TestKeywords_LowercaseOutput(t *testing.T) {
	r := newResume("Docker KUBERNETES Docker")

	got := r.Keywords()

	for _, tok := range got {
		require.Equal(t, tok, strings.ToLower(tok), "token %q is not lower-case", tok)
	}
}

func TestKeywords_DeduplicatesTokens(t *testing.T) {
	r := newResume("Docker docker DOCKER")

	got := r.Keywords()

	require.ElementsMatch(t,
		[]string{"docker", "golang", "concurrency", "sql"},
		got,
		"expected duplicates to be removed",
	)
}
