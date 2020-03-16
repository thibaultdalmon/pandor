package databases

import (
	"encoding/json"
	"fmt"
	"log"
	"pandor/logger"
	"pandor/models"
	"testing"

	"github.com/dgraph-io/dgo/v2/protos/api"
)

// Format allows to extract the information from a Response
func Format(resp *api.Response) error {
	type Root struct {
		Me []models.Author `json:"test"`
	}

	var r Root
	err := json.Unmarshal(resp.Json, &r)
	if err != nil {
		logger.Logger.Fatal(err.Error())
	}

	out, _ := json.MarshalIndent(r, "", "\t")
	fmt.Printf("%s\n", out)
	fmt.Println(string(resp.Json))

	return err
}

func TestDB(t *testing.T) {
	d, dg, err := NewClient()
	if err != nil {
		log.Fatal(err)
	}
	defer d.Close()

	err = DropAll(dg)
	if err != nil {
		log.Fatal(err)
	}

	err = LoadSchema(models.Schema, dg)
	if err != nil {
		log.Fatal(err)
	}

	uidarticle := models.FormatUID("Another Article from the same guy")
	uidauthor := models.FormatUID("Lewis_G")
	sub := models.FormatTime("2007-12-29T00:00:00.000Z")
	crw := models.FormatTime("2020-03-09T08:44:03.484Z")
	article := models.Article{
		UID:            uidarticle,
		ArXivID:        "0801.0003",
		Title:          "Another Article from the same guy",
		Abstract:       `Cluster Stuff`,
		SubmissionDate: sub,
		CrawledAt:      crw,
		PDFURL:         "https://export.arxiv.org/pdf/0801.0003",
		OtherFormatURL: "https://export.arxiv.org/format/0801.0003",
		MetaURL:        "https://export.arxiv.org/abs/0801.00003",
		Authors: []models.Author{
			{
				UID:   uidauthor,
				Name:  "Lewis_G",
				URL:   "https://export.arxiv.org/find/astro-ph/1/au:+Lewis_G/0/1/0/all/0/1",
				DType: []string{"Author"},
			},
		},
		DType: []string{"Article"},
	}

	_, err = AddArticle(article, dg)
	if err != nil {
		log.Fatal(err)
	}

	uidarticle = models.FormatUID("Globular clusters in the outer halo of M31: the survey")
	uidauthor, err = GetAuthorUID("Lewis_G", dg)
	if err != nil {
		log.Fatal(err)
	}
	sub = models.FormatTime("2007-12-28T00:00:00.000Z")
	crw = models.FormatTime("2020-03-08T08:44:03.484Z")
	article = models.Article{
		UID:     uidarticle,
		ArXivID: "0801.0002",
		Title:   "Globular clusters in the outer halo of M31: the survey",
		Abstract: `We report the discovery of 40 new globular clusters (GCs)
		that have been\nfound in surveys of the halo of M31 based on INT/WFC and
		CHFT/Megacam imagery.\nA subset of these these new GCs are of an extended,
		diffuse nature, and include\nthose already found in Huxor et al. (2005).
		The search strategy is described\nand basic positional and V and I
		photometric data are presented for each\ncluster. For a subset of these
		clusters, K-band photometry is also given. The\nnew clusters continue to
		be found to the limit of the survey area (~100 kpc),\nrevealing that the GC
		system of M31 is much more extended than previously\nrealised. The new
		clusters increase the total number of confirmed GCs in M31 by\napproximately
		10% and the number of confirmed GCs beyond 1 degree (~14 kpc) by\nmore than
		75%. We have also used the survey imagery as well recent HST archival\ndata
		to update the Revised Bologna Catalogue (RBC) of M31 globular clusters.`,
		SubmissionDate: sub,
		CrawledAt:      crw,
		PDFURL:         "https://export.arxiv.org/pdf/0801.0002",
		OtherFormatURL: "https://export.arxiv.org/format/0801.0002",
		MetaURL:        "https://export.arxiv.org/abs/0801.00002",
		Authors: []models.Author{
			{
				UID:   uidauthor,
				Name:  "Lewis_G",
				URL:   "https://export.arxiv.org/find/astro-ph/1/au:+Lewis_G/0/1/0/all/0/1",
				DType: []string{"Author"},
			},
		},
		DType: []string{"Article"},
	}

	_, err = AddArticle(article, dg)
	if err != nil {
		log.Fatal(err)
	}

	_, err = AddArticle(article, dg)
	if err != nil {
		log.Fatal(err)
	}

	query := `{
		test(func: eq(url, "https://export.arxiv.org/find/astro-ph/1/au:+Lewis_G/0/1/0/all/0/1")){
		uid
    name
    url
    ~authors{
      title
    }
  }
}`

	var resp api.Response
	resp, err = Query(query, dg)
	if err != nil {
		log.Fatal(err)
	}

	err = Format(&resp)
	if err != nil {
		log.Fatal(err)
	}
}
