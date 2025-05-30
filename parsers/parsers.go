package parsers

type Parser interface {
	ParseJob(url string) (*Job, error)
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
