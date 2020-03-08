package scrappers

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/thibaultdalmon/pandor/models"

	"code.sajari.com/docconv"
	"github.com/gocolly/colly"
	"github.com/gocolly/colly/debug"
	"github.com/gocolly/colly/queue"
	"go.uber.org/zap"
)

// Domain is the domain name
var Domain = "https://export.arxiv.org"

// TempDir is the directory to store temporary files
var TempDir = "./tmp/"

// ParsePDF converts a PDF to a string
func ParsePDF(inputPath string) error {
	res, err := docconv.ConvertPath(inputPath)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(res)

	return err
}

// exists returns whether the given file or directory exists
func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return true, err
}

// DownloadAndSaveToDir is an helper function to save the result of a GET Request
func DownloadAndSaveToDir(url, file, dir string) error {
	// Create the file
	if dirExists, err := exists(dir); dirExists {
		if err != nil {
			return err
		}
	} else {
		err = os.MkdirAll(dir, 0777)
		if err != nil {
			return err
		}
	}

	var out *os.File
	if dirExists, err := exists(dir + file); dirExists {
		if err != nil {
			return err
		}
	} else {
		out, err = os.Create(dir + file)
		defer out.Close()
		if err != nil {
			return err
		}
	}

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

// LaunchArXiv creates an ArXiv web crawler and runs it
func LaunchArXiv() {
	url := Domain + "/abs/"

	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("can't initialize zap logger: %v", err)
	}
	defer logger.Sync()

	// create a request queue with 2 consumer threads
	var q *queue.Queue
	q, err = queue.New(
		4, // Number of consumer threads
		&queue.InMemoryQueueStorage{MaxSize: 10000}, // Use default queue storage
	)
	if err != nil {
		logger.Fatal(fmt.Sprintf("can't initialize queue: %v", err))
	}

	// Instantiate default collector
	c := colly.NewCollector(
		colly.Debugger(&debug.LogDebugger{}),
		// Visit only domains: export.arxiv.org
		colly.AllowedDomains("export.arxiv.org"),
		colly.Async(true),
	)

	// Limit the maximum parallelism to 2
	// This is necessary if the goroutines are dynamically
	// created to control the limit of simultaneous requests.
	//
	// Parallelism can be controlled also by spawning fixed
	// number of go routines.
	c.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		RandomDelay: 1 * time.Second,
		Parallelism: 4,
	})

	c.OnHTML(`div[id=abs]`, func(e *colly.HTMLElement) {
		article := models.Article{}

		article.HTMLResponse = string(e.Response.Body)

		article.URL = e.Request.Ctx.Get("url")

		article.CrawledAt = time.Now().UTC()

		article.Title = strings.SplitAfterN(e.ChildText(`h1.title`), "\n", 2)[1]

		article.Abstract = strings.SplitAfterN(
			e.ChildText(`blockquote.abstract`),
			" ",
			2)[1]

		// Authors
		authors := e.ChildTexts(`div.authors a`)
		article.Authors = make([]string, len(authors))
		copy(article.Authors, authors)

		// SubmissionDate
		SubmissionDateStr := e.ChildText(`div.dateline`)
		re := regexp.MustCompile(`\d{2}\s\w{3}\s\d{4}`)
		if re.MatchString(SubmissionDateStr) {
			SubmissionDateStrFmted := re.FindString(SubmissionDateStr)
			SubmissionDateT, err := time.Parse("2 Jan 2006", SubmissionDateStrFmted)
			if err == nil {
				article.SubmissionDate = SubmissionDateT
			} else {
				logger.Error(fmt.Sprintf("SubmissionDate Parsing Error: %v", err))
			}
		}

		// Article Links
		if attr, ok := e.DOM.Find(`div.full-text li a`).First().Attr(`href`); ok {
			article.PDF = Domain + attr
		}
		re = regexp.MustCompile(`\d{4}.((\d{5}$)|(\d{4}$))`)
		filePath := re.FindString(article.PDF) + ".pdf"
		err = DownloadAndSaveToDir(article.PDF, filePath, TempDir)
		if err != nil {
			logger.Error(fmt.Sprintf("FILE Error %v", err))
		}

		if attr, ok := e.DOM.Find(`div.full-text li a`).Last().Attr(`href`); ok {
			article.Format = Domain + attr
		}

		logger.Debug("New Article",
			zap.String("URL:", article.URL),
			zap.Time("CrawledAt:", article.CrawledAt),
			zap.String("Title:", article.Title),
			zap.String("Abstract:", article.Abstract),
			zap.Int("Nb Authors:", len(article.Authors)),
			zap.String("Last Author:", article.Authors[len(article.Authors)-1]),
			zap.Time("Submission Date:", article.SubmissionDate),
			zap.String("PDF:", article.PDF),
			zap.String("Format:", article.Format),
		)
	})

	// Before making a request print "Visiting ..."
	c.OnRequest(func(r *colly.Request) {
		r.Ctx.Put("url", r.URL.String())
		logger.Info(fmt.Sprintf("Visiting %s", r.URL.String()))
	})
	// Set error handler
	c.OnError(func(r *colly.Response, err error) {
		logger.Error(fmt.Sprintf("Request URL: %s failed with response: %v", r.Request.URL, r),
			zap.String("Error:", fmt.Sprintf("%v", err)),
		)
	})
	// Called after OnHTML
	c.OnScraped(func(r *colly.Response) {
		logger.Info(fmt.Sprintf("Finished %s", r.Request.URL))
		URL := r.Request.URL.String()
		Base := regexp.MustCompile(`.*\d{4}\.`)
		URLBase := Base.FindString(URL)
		Number := regexp.MustCompile(`(\d{5}$)|(\d{4}$)`)
		ArticleNumber, err := strconv.Atoi(Number.FindString(URL))
		if err != nil {
			logger.Error(fmt.Sprintf("Error: %v", err))
		}
		if ArticleNumber < 2 {
			// Visit next article
			r.Request.Visit(fmt.Sprintf("%s%05d", URLBase, ArticleNumber+1))
			logger.Info(fmt.Sprintf("Adding %s%05d", URLBase, ArticleNumber+1))
		}
	})
	// Start scraping
	for i := 8; i < 9; i++ {
		for j := 1; j < 2; j++ {
			// Add URLs to the queue
			q.AddURL(fmt.Sprintf("%s%02d%02d.00001", url, i, j))
		}
	}
	q.Run(c)
	// Wait until threads are finished
	c.Wait()

}
