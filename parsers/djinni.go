package parsers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/mmcdole/gofeed"
)

type Djinni struct {
	c *http.Client
}

func NewDjinni() Djinni {
	return Djinni{
		c: &http.Client{},
	}
}

func (d Djinni) client() *http.Client {
	return d.c
}

func (Djinni) Host() string {
	return "djinni.co"
}

func (d Djinni) ParseJob(url string) (*Job, error) {
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

	job := d.findJob(doc)
	if job == nil {
		return nil, errors.New("could not find job JSON-LD data")
	}

	return &Job{
		URL:            job.URL,
		Title:          job.Title,
		Description:    strings.TrimSpace(job.Description),
		DatePosted:     job.DatePosted,
		EmploymentType: job.EmploymentType,
		Industry:       job.Industry,
		LocationType:   job.JobLocationType,
		ValidThrough:   job.ValidThrough,
	}, nil
}

type jobJSONLD struct {
	Title           string `json:"title"`
	Description     string `json:"description"`
	URL             string `json:"url"`
	DatePosted      string `json:"datePosted"`
	EmploymentType  string `json:"employmentType"`
	Industry        string `json:"industry"`
	JobLocationType string `json:"jobLocationType"`
	ValidThrough    string `json:"validThrough"`
}

func (Djinni) findJob(doc *goquery.Document) (job *jobJSONLD) {
	selector := func(_ int, s *goquery.Selection) bool {
		raw := s.Text()
		if !strings.Contains(raw, `"@type": "JobPosting"`) {
			return true // continue
		}

		err := json.Unmarshal([]byte(raw), job)
		found := err == nil && job.Title != ""
		if found {
			return false // break
		}

		return true
	}

	job = new(jobJSONLD)
	doc.Find("script[type='application/ld+json']").EachWithBreak(selector)
	return job
}

func (d Djinni) ParseFeed() ([]FeedEntry, error) {
	const url = "https://djinni.co/jobs/rss/?company_type=outsource&company_type=outstaff&company_type=product&company_type=startup"

	p := gofeed.NewParser()
	p.Client = d.client()

	f, err := p.ParseURL(url)
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

func (Djinni) normalizeFeedEntry(fe FeedEntry) FeedEntry {
	fe.ID = strings.TrimPrefix(fe.ID, "https://djinni.co/")
	fe.ID = strings.TrimSuffix(fe.ID, "/")
	fe.Description = strings.TrimSpace(fe.Description)
	return fe
}
