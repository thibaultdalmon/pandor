package models

import "time"

// Article type
type Article struct {
	Title          string
	Abstract       string
	SubmissionDate time.Time
	CrawledAt      time.Time
	HTMLResponse   string
	PDF            string
	Format         string
	URL            string
	Authors        []string
	CitedPapers    []string
	CitedBy        []string
}
