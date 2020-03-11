package scrappers

import (
	"fmt"
	"log"
	"testing"
)

func TestNameExtraction(t *testing.T) {
	url := "https://export.arxiv.org/find/astro-ph/1/au:+Lewis_G/0/1/0/all/0/1"
	name, err := ExtractNameFromURL(url)
	if err != nil {
		log.Fatal(err)
	}
	if name != "Lewis_G" {
		log.Fatal(fmt.Errorf("Wrong name: %s instead of Lewis_G", name))
	}
}
