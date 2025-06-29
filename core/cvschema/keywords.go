package cvschema

import (
	"regexp"
	"strings"

	"github.com/bbalet/stopwords"
	"github.com/kljensen/snowball"
)

// String returns a string representation of the Resume.
// Implements fmt.Stringer interface.
func (r Resume) String() string {
	var b strings.Builder
	b.WriteString(r.Basics.Label + " ")
	b.WriteString(r.Basics.Summary + " ")

	for _, s := range r.Skills {
		b.WriteString(s.Name + " ")
		for _, kw := range s.Keywords {
			b.WriteString(kw + " ")
		}
	}

	for _, w := range r.Work {
		b.WriteString(w.Position + " ")
		b.WriteString(w.Summary + " ")
		b.WriteString(w.Description + " ")
		for _, h := range w.Highlights {
			b.WriteString(h + " ")
		}
	}

	for _, p := range r.Projects {
		b.WriteString(p.Name + " ")
		b.WriteString(p.Description + " ")
		for _, h := range p.Highlights {
			b.WriteString(h + " ")
		}
		for _, kw := range p.Keywords {
			b.WriteString(kw + " ")
		}
	}

	for _, e := range r.Education {
		b.WriteString(e.Area + " ")
		for _, c := range e.Courses {
			b.WriteString(c + " ")
		}
	}

	return b.String()
}

// Keywords returns lower-case keywords based on the resume schema.
func (r Resume) Keywords() []string {
	text := strings.TrimSpace(r.String())

	// 1. Remove stop-words
	text = stopwords.CleanString(text, "en", true)

	// 2. Tokenize (unicode-aware)
	words := regexp.MustCompile(`[a-z0-9]+`).FindAllString(text, -1)

	// 3. De-duplicate
	seen := make(map[string]bool, len(words))

	out := make([]string, 0, len(words))
	for _, w := range words {
		lw := strings.ToLower(w)

		// 4. Strip short tokens
		if len(lw) <= 2 {
			continue
		}

		// 5. Cut a word down to its root
		stem, err := snowball.Stem(w, "english", true)
		if err == nil && stem != "" {
			w = stem
		}

		if !seen[w] {
			seen[w] = true
			out = append(out, w)
		}
	}

	return out
}
