package utils

import (
	"fmt"
	"log"

	"code.sajari.com/docconv"
)

// ParsePDF converts a PDF to a string
func ParsePDF(inputPath string) error {
	res, err := docconv.ConvertPath(inputPath)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(res)

	return err
}
