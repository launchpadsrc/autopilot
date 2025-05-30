package parsers

import (
	"cmp"
	"fmt"
	"io"
	"net/http"
	"slices"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type Dou struct {
	client *http.Client
}

func NewDou() Dou {
	return Dou{
		client: &http.Client{},
	}
}

func (d Dou) ParseJob(url string) (*Job, error) {
	resp, err := d.client.Get(url)
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
