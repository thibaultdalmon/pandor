package scrappers

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"pandor/logger"
	"pandor/models"
	"pandor/utils"

	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/debug"
	"github.com/gocolly/colly/v2/queue"
	"go.uber.org/zap"
)

// Domain is the domain name
var Domain = "https://export.arxiv.org"

// TempDir is the directory to store temporary files
var TempDir = "./tmp/"

// LaunchArXiv creates an ArXiv web crawler and runs it
func LaunchArXiv() {
	url := Domain + "/abs/"

	// create a request queue with 2 consumer threads
	q, err := queue.New(
		4, // Number of consumer threads
		&queue.InMemoryQueueStorage{MaxSize: 10000}, // Use default queue storage
	)
	if err != nil {
		logger.Logger.Fatal(fmt.Sprintf("can't initialize queue: %v", err))
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

		article.MetaURL = e.Request.Ctx.Get("url")

		crawlingTime := time.Now().UTC()
		article.CrawledAt = crawlingTime

		article.Title = strings.SplitAfterN(e.ChildText(`h1.title`), "\n", 2)[1]

		article.Abstract = strings.SplitAfterN(
			e.ChildText(`blockquote.abstract`),
			" ",
			2)[1]

		// Authors
		authors := e.ChildTexts(`div.authors a`)
		article.Authors = make([]models.Author, len(authors))
		for author := 0; author < len(authors); author++ {
			article.Authors[author].Name = authors[author]
		}

		// SubmissionDate
		SubmissionDateStr := e.ChildText(`div.dateline`)
		re := regexp.MustCompile(`\d{2}\s\w{3}\s\d{4}`)
		var SubmissionDateT time.Time
		if re.MatchString(SubmissionDateStr) {
			SubmissionDateStrFmted := re.FindString(SubmissionDateStr)
			SubmissionDateT, err = time.Parse("2 Jan 2006", SubmissionDateStrFmted)
			if err == nil {
				article.SubmissionDate = SubmissionDateT
			} else {
				logger.Logger.Error(fmt.Sprintf("SubmissionDate Parsing Error: %v", err))
			}
		}

		// Article Links
		if attr, ok := e.DOM.Find(`div.full-text li a`).First().Attr(`href`); ok {
			article.PDFURL = Domain + attr
		}
		re = regexp.MustCompile(`\d{4}.((\d{5}$)|(\d{4}$))`)
		filePath := re.FindString(article.PDFURL) + ".pdf"
		err = utils.DownloadAndSaveToDir(article.PDFURL, filePath, TempDir)
		if err != nil {
			logger.Logger.Error(fmt.Sprintf("FILE Error: %v", err))
		}

		if attr, ok := e.DOM.Find(`div.full-text li a`).Last().Attr(`href`); ok {
			article.OtherFormatURL = Domain + attr
		}

		logger.Logger.Debug("New Article",
			zap.String("URL:", article.MetaURL),
			zap.Time("CrawledAt:", article.CrawledAt),
			zap.String("Title:", article.Title),
			zap.String("Abstract:", article.Abstract),
			zap.Int("Nb Authors:", len(article.Authors)),
			zap.String("Last Author:", article.Authors[len(article.Authors)-1].Name),
			zap.Time("Submission Date:", article.SubmissionDate),
			zap.String("PDF:", article.PDFURL),
			zap.String("Format:", article.OtherFormatURL),
		)
	})

	// Before making a request print "Visiting ..."
	c.OnRequest(func(r *colly.Request) {
		r.Ctx.Put("url", r.URL.String())
		logger.Logger.Info(fmt.Sprintf("Visiting %s", r.URL.String()))
	})
	// Set error handler
	c.OnError(func(r *colly.Response, err error) {
		logger.Logger.Error(fmt.Sprintf("Request URL: %s failed with response: %v", r.Request.URL, r),
			zap.String("Error:", fmt.Sprintf("%v", err)),
		)
	})
	// Called after OnHTML
	c.OnScraped(func(r *colly.Response) {
		logger.Logger.Info(fmt.Sprintf("Finished %s", r.Request.URL))
		URL := r.Request.URL.String()
		Base := regexp.MustCompile(`.*\d{4}\.`)
		URLBase := Base.FindString(URL)
		Number := regexp.MustCompile(`(\d{5}$)|(\d{4}$)`)
		ArticleNumber, err := strconv.Atoi(Number.FindString(URL))
		if err != nil {
			logger.Logger.Error(fmt.Sprintf("Error: %v", err))
		}
		if ArticleNumber < 2 {
			// Visit next article
			r.Request.Visit(fmt.Sprintf("%s%05d", URLBase, ArticleNumber+1))
			logger.Logger.Info(fmt.Sprintf("Adding %s%05d", URLBase, ArticleNumber+1))
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
