package main

import (
	"pandor/logger"
	"pandor/scrappers"
)

func main() {
	logger.Logger = logger.InitLogger()
	defer logger.Logger.Sync()
	scrappers.LaunchArXiv()
}
