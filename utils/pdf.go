package utils

import (
	"fmt"

	"code.sajari.com/docconv"
)

// PDFtoTXT converts a PDF to a string
func PDFtoTXT(inputPath string) (*docconv.Response, error) {
	res, err := docconv.ConvertPath(inputPath)
	fmt.Println(res.Error)
	return res, err
}
