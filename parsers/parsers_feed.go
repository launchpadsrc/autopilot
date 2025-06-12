package parsers

import (
	"strings"
	"time"
	"unicode/utf8"

	"github.com/PuerkitoBio/goquery"
)

type FeedEntry struct {
	ID          string
	Link        string
	Title       string
	Description string
	Published   *time.Time
}

// BoltID implements `bboltx.BucketValue`.
func (fe FeedEntry) BoltID() string {
	return fe.ID
}

// FirstParagraphs extracts the first paragraph from the feed entry's HTML
// description. If the first paragraph is shorter than 120 runes, it keeps
// appending the following <p> elements (in order) until the total length
// is at least 120 runes, or there are no more paragraphs left.
func (fe FeedEntry) FirstParagraphs() string {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(fe.Description))
	if err != nil {
		return fe.Description
	}

	ps := doc.Find("p")
	if ps.Length() == 0 {
		return fe.Description
	}

	var b strings.Builder
	ps.EachWithBreak(func(i int, s *goquery.Selection) bool {
		text := strings.TrimSpace(s.Text())
		if text == "" {
			return true // continue
		}

		if b.Len() > 0 {
			b.WriteString("\n\n") // blank line between paragraphs
		}
		b.WriteString(text)

		// Stop iterating once we've reached the desired length.
		return utf8.RuneCountInString(b.String()) < 120
	})

	return b.String()
}
