package models

import "time"

// Article type
// If omitempty is not set, then edges with empty values (0 for int/float, "" for string, false
// for bool) would be created for values not specified explicitly.
type Article struct {
	UID            string    `json:"uid,omitempty"`
	ArXivID        string    `json:"arxivid,omitempty"`
	Title          string    `json:"title,omitempty"`
	Abstract       string    `json:"abstract,omitempty"`
	SubmissionDate time.Time `json:"submissiondate,omitempty"`
	CrawledAt      time.Time `json:"crawledat,omitempty"`
	HTMLResponse   string    `json:"htmlresponse,omitempty"`
	PDFURL         string    `json:"pdfurl,omitempty"`
	OtherFormatURL string    `json:"otherformaturl,omitempty"`
	MetaURL        string    `json:"metaurl,omitempty"`
	Authors        []Author  `json:"authors,omitempty"`
	CitedPapers    []Article `json:"citedpapers,omitempty"`
	DType          []string  `json:"dgraph.type,omitempty"`
}

// Author type
type Author struct {
	UID   string   `json:"uid,omitempty"`
	Name  string   `json:"name,omitempty"`
	URL   string   `json:"url,omitempty"`
	DType []string `json:"dgraph.type,omitempty"`
}

// Schema describing the types
var Schema = `
  title: string @index(term, exact, hash, fulltext, trigram) .
  name: string @index(term, exact, hash, fulltext, trigram) .
	url: string @index(hash) .
	arxivid: string @index(term, exact, hash, fulltext, trigram) .
  abstract: string .
  submissiondate: datetime .
  crawledat: datetime .
  htmlresponse: string .
  pdfurl: string .
  otherformaturl: string .
  metaurl: string .
  authors: [uid] @reverse .
  citedpapers: [uid] @reverse .

  type Article {
		arxivid: string
    title: string
    abstract: string
    submissiondate: datetime
    crawledat: datetime
    htmlresponse: string
    pdfurl: string
    otherformaturl: string
    metaurl: string
    authors: [Author]
    citedpapers: [Article]
  }

  type Author {
    name: string
		url: string
  }
`
