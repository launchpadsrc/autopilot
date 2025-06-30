package parsers

import (
	"cmp"
	"fmt"
	"io"
	"net/http"
	"slices"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/mmcdole/gofeed"
)

type Dou struct {
	c *http.Client
}

func NewDou() Dou {
	return Dou{
		c: &http.Client{},
	}
}

func (d Dou) client() *http.Client {
	return d.c
}

func (Dou) Host() string {
	return "jobs.dou.ua"
}

func (d Dou) ParseJob(url string) (*Job, error) {
	resp, err := d.c.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch URL: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read body: %w", err)
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
	if err != nil {
		return nil, fmt.Errorf("failed to parse html: %w", err)
	}

	var (
		canonicalLink, _ = doc.Find("link[rel='canonical']").Attr("href")
		title            = strings.TrimSpace(doc.Find("h1.g-h2").First().Text())
		description      = strings.TrimSpace(doc.Find(".b-typo.vacancy-section").First().Text())
		datePosted       = strings.TrimSpace(doc.Find(".l-vacancy .date").First().Text())
		location         = strings.TrimSpace(doc.Find(".sh-info .place").First().Text())
		locationType     string
	)

	switch {
	case slices.Contains(locationsInUA, strings.ToLower(location)):
		locationType = "REMOTE"
	default:
		locationType = "ONSITE"
	}

	return &Job{
		URL:          cmp.Or(canonicalLink, url),
		Title:        title,
		Description:  description,
		DatePosted:   datePosted,
		LocationType: locationType,
	}, nil
}

var locationsInUA = []string{"віддалено", "за кордоном"}

func (d Dou) ParseFeed() ([]FeedEntry, error) {
	const url = "https://jobs.dou.ua/vacancies/feeds/"

	f, err := gofeed.NewParser().ParseURL(url)
	if err != nil {
		return nil, err
	}

	entries := make([]FeedEntry, 0, len(f.Items))
	for _, it := range f.Items {
		id := it.GUID
		if id == "" {
			id = it.Link // fallback if GUID missing
		}

		fe := FeedEntry{
			ID:          id,
			Title:       it.Title,
			Link:        it.Link,
			Published:   it.PublishedParsed,
			Description: it.Description,
		}

		entries = append(entries, d.normalizeFeedEntry(fe))
	}

	return entries, nil
}

func (Dou) normalizeFeedEntry(fe FeedEntry) FeedEntry {
	fe.ID = strings.TrimPrefix(fe.ID, "https://jobs.dou.ua/")
	fe.ID = strings.Split(fe.ID, "/?")[0]
	fe.Link = strings.TrimSuffix(fe.Link, "/?utm_source=jobsrss")
	fe.Description = strings.TrimPrefix(fe.Description, "Project description:")
	fe.Description = strings.TrimSuffix(fe.Description, "Відгукнутися на вакансію")
	fe.Description = strings.TrimSpace(fe.Description)
	return fe
}
