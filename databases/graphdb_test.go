package databases

import (
	"log"
	"pandor/models"
	"strings"
	"testing"
	"time"
)

func makeUID(s string) string {
	uid := "_:" + strings.ToLower(s)
	uid = strings.ReplaceAll(uid, " ", "")
	return uid
}

func makeTime(s string) time.Time {
	t, _ := time.Parse("2006-01-02T15:04:05.000Z", s)
	return t
}

func TestDB(t *testing.T) {
	dg, err := NewClient()
	if err != nil {
		log.Fatal(err)
	}

	err = DropAll(dg)
	if err != nil {
		log.Fatal(err)
	}

	err = LoadSchema(models.Schema, dg)
	if err != nil {
		log.Fatal(err)
	}

	uidarticle := makeUID("Globular clusters in the outer halo of M31: the survey")
	uidauthor := makeUID("G. F. Lewis")
	sub := makeTime("2007-12-28T00:00:00.000Z")
	crw := makeTime("2020-03-08T08:44:03.484Z")
	article := models.Article{
		UID:   uidarticle,
		Title: "Globular clusters in the outer halo of M31: the survey",
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
				Name:  "G. F. Lewis",
				DType: []string{"Author"},
			},
		},
		CitedPapers: []models.Article{},
		CitedBy:     []models.Article{},
		DType:       []string{"Article"},
	}

	response, err := AddArticle(article, dg)
	if err != nil {
		log.Fatal(err)
	}

	log.Println(response.Uids)

	variables := map[string]string{"$id1": response.Uids["globularclustersintheouterhaloofm31:thesurvey"]}
	query := `query PrefArt($id1: string){
    art(func: uid($id1)) {
      title
      abstract
			submissiondate
      crawledat
      pdfurl
      dgraph.type
      authors @filter(eq(name, "G. F. Lewis")){
        name
        dgraph.type
      }
    }
  }`

	resp, err := Query(query, variables, dg)
	if err != nil {
		log.Fatal(err)
	}

	err = Format(resp)
	if err != nil {
		log.Fatal(err)
	}
}
