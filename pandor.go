package main

import (
	"fmt"

	"pandor/logger"
	"pandor/scrappers"
)

func main() {
	fmt.Println("Hello World")
	logger.Logger = logger.InitLogger()
	defer logger.Logger.Sync()
	scrappers.LaunchArXiv()
}
