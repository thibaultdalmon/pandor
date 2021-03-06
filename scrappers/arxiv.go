package scrappers

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"pandor/databases"
	"pandor/logger"
	"pandor/models"

	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/queue"
	"go.uber.org/zap"
)

// Domain is the domain name
var Domain = "https://export.arxiv.org"

// TempDir is the directory to store temporary files
var TempDir = "./tmp/"

// ExtractNameFromURL extracts the name of an author from its URL
func ExtractNameFromURL(url string) (string, error) {
	re := regexp.MustCompile(`^.*\+`)
	if !re.MatchString(url) {
		re = regexp.MustCompile(`javascript`)
		if !re.MatchString(url) {
			logger.Logger.Fatal(fmt.Sprintf("Prefix not Found: %s ", url))
		}
		return "", fmt.Errorf("Prefix not Found: %s ", url)
	}
	prefix := re.FindString(url)
	re = regexp.MustCompile(`/.*$`)
	if !re.MatchString(url) {
		re = regexp.MustCompile(`javascript`)
		if !re.MatchString(url) {
			logger.Logger.Fatal(fmt.Sprintf("Prefix not Found: %s ", url))
		}
		return "", fmt.Errorf("Prefix not Found: %s ", url)
	}
	url = strings.Replace(url, prefix, "", 1)
	suffix := re.FindString(url)
	url = strings.Replace(url, suffix, "", 1)

	return url, nil
}

// LaunchArXiv creates an ArXiv web crawler and runs it
func LaunchArXiv() {
	url := Domain + "/abs/"

	conn, dg, err := databases.NewClient()
	if err != nil {
		logger.Logger.Fatal(err.Error())
	}
	defer conn.Close()

	// err = databases.DropAll(dg)
	// if err != nil {
	// 	logger.Logger.Error(err.Error())
	// }

	err = databases.LoadSchema(models.Schema, dg)
	if err != nil {
		logger.Logger.Fatal(err.Error())
	}

	// create a request queue with 2 consumer threads
	q, err := queue.New(
		4, // Number of consumer threads
		&queue.InMemoryQueueStorage{MaxSize: 100000}, // Use default queue storage
	)
	if err != nil {
		logger.Logger.Fatal(fmt.Sprintf("can't initialize queue: %v", err))
	}

	// Instantiate default collector
	c := colly.NewCollector(
		// colly.Debugger(&debug.LogDebugger{}),
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
		Parallelism: 8,
	})

	c.OnHTML(`div[id=abs]`, func(e *colly.HTMLElement) {

		conn, dg, err := databases.NewClient()
		if err != nil {
			logger.Logger.Error(err.Error())
		}
		defer conn.Close()

		article := models.Article{}

		article.HTMLResponse = string(e.Response.Body)

		article.MetaURL = e.Request.URL.String()
		re := regexp.MustCompile(`\d{4}.(\d{5}|\d{4})`)
		if !re.MatchString(article.MetaURL) {
			logger.Logger.Error(fmt.Sprintf("No ID Matched for %s", article.MetaURL))
		}
		article.ArXivID = re.FindString(article.MetaURL)

		crawlingTime := time.Now().UTC()
		article.CrawledAt = crawlingTime

		article.Title = strings.SplitAfterN(e.ChildText(`h1.title`), "\n", 2)[1]
		article.UID = models.FormatUID(article.Title)
		article.DType = []string{"Article"}

		article.Abstract = strings.SplitAfterN(
			e.ChildText(`blockquote.abstract`),
			" ",
			2)[1]

		// Authors
		authorsURL := e.ChildAttrs(`div.authors a`, `href`)
		article.Authors = make([]models.Author, len(authorsURL))
		for author := 0; author < len(authorsURL); author++ {
			authorURL := authorsURL[author]
			name, err := ExtractNameFromURL(authorURL)
			if err != nil {
				continue
			}
			uid, err := databases.GetAuthorUID(name, dg)
			if err != nil {
				article.Authors[author].UID = models.FormatUID(name)
			} else {
				article.Authors[author].UID = uid
			}
			article.Authors[author].URL = Domain + authorURL
			article.Authors[author].Name = name
			article.Authors[author].DType = []string{"Author"}
		}

		// SubmissionDate
		SubmissionDateStr := e.ChildText(`div.dateline`)
		re = regexp.MustCompile(`\d{2}\s\w{3}\s\d{4}`)
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
		// re = regexp.MustCompile(`\d{4}.((\d{5}$)|(\d{4}$))`)
		// filePath := re.FindString(article.PDFURL) + ".pdf"
		// err = utils.DownloadAndSaveToDir(article.PDFURL, filePath, TempDir)
		// if err != nil {
		// 	logger.Logger.Error(fmt.Sprintf("FILE Error: %v", err))
		// }

		if attr, ok := e.DOM.Find(`div.full-text li a`).Last().Attr(`href`); ok {
			article.OtherFormatURL = Domain + attr
		}

		_, err = databases.AddArticle(article, dg)
		if err != nil {
			logger.Logger.Error(err.Error())
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

		conn, dg, err := databases.NewClient()
		if err != nil {
			logger.Logger.Error(err.Error())
		}
		defer conn.Close()

		logger.Logger.Info(fmt.Sprintf("Finished %s", r.Request.URL))
		URL := r.Request.URL.String()

		Base := regexp.MustCompile(`.*\d{4}\.`)
		if !Base.MatchString(URL) {
			logger.Logger.Error(fmt.Sprintf("Wrong base in URL: %s", URL))
		}
		URLBase := Base.FindString(URL)

		Date := regexp.MustCompile(`\d{4}\.$`)
		if !Date.MatchString(URLBase) {
			logger.Logger.Error(fmt.Sprintf("Wrong date in URL: %s", URLBase))
		}
		URLDate := Date.FindString(URLBase)

		Number := regexp.MustCompile(`(\d{5}|\d{4})$`)
		if !Number.MatchString(URL) {
			logger.Logger.Error(fmt.Sprintf("Wrong number in URL: %s", URL))
		}
		URLNumber := Number.FindString(URL)

		ArticleNumber, err := strconv.Atoi(URLNumber)
		if err != nil {
			logger.Logger.Error(fmt.Sprintf("Error: %v", err))
		}

		type Root struct {
			Articles []models.Article `json:"getuid"`
		}
		var root Root

		for {
			ArticleNumber++
			variables := map[string]string{"$id": fmt.Sprintf("%s%05d", URLDate, ArticleNumber)}
			query := `query GetUID($id: string){
									getuid(func: eq(arxivid, $id)){
										uid
								  }
								}`
			resp, err := databases.QueryWithVars(query, variables, dg)
			if err != nil {
				logger.Logger.Fatal(err.Error())
			}
			err = json.Unmarshal(resp.Json, &root)
			if err != nil {
				logger.Logger.Fatal(err.Error())
			}

			if len(root.Articles) != 0 {
				continue
			}

			variables = map[string]string{"$id": fmt.Sprintf("%s%04d", URLDate, ArticleNumber)}
			resp, err = databases.QueryWithVars(query, variables, dg)
			if err != nil {
				logger.Logger.Fatal(err.Error())
			}
			err = json.Unmarshal(resp.Json, &root)
			if err != nil {
				logger.Logger.Fatal(err.Error())
			}

			if len(root.Articles) == 0 {
				break
			}
		}
		if ArticleNumber < 100 {
			// Visit next article
			r.Request.Visit(fmt.Sprintf("%s%05d", URLBase, ArticleNumber))
			logger.Logger.Info(fmt.Sprintf("Adding %s%05d", URLBase, ArticleNumber))
		}
	})

	// Start scraping
	for i := 8; i < 21; i++ {
		for j := 1; j < 13; j++ {
			// Add URLs to the queue
			q.AddURL(fmt.Sprintf("%s%02d%02d.00001", url, i, j))
		}
	}
	q.Run(c)
	// Wait until threads are finished
	c.Wait()
}
