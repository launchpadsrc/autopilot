package parsers

type Parser interface {
	Host() string
	ParseJob(url string) (*Job, error)
	ParseFeed() ([]FeedEntry, error)
}

type Job struct {
	URL            string
	Title          string
	Description    string
	DatePosted     string
	EmploymentType string
	Industry       string
	LocationType   string
	ValidThrough   string
}
