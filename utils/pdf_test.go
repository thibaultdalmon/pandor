package utils

import (
	"fmt"
	"log"
	"testing"
)

func TestPDF(t *testing.T) {
	path := "../tmp/0801.0001.pdf"
	res, err := PDFtoTXT(path)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(res.Body)
}
