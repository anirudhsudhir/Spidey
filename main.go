package main

import (
	"fmt"
	"time"

	"github.com/anirudhsudhir/spidey/crawl"
)

func main() {
	seedUrls := []string{}
	crawlStats := crawl.CrawlLinks(seedUrls, time.Second)
	fmt.Printf("TotalCrawls: %d, SuccessfulCrawls: %d, FailedCrawls: %d", crawlStats.TotalCrawls, crawlStats.SuccessfulCrawls, crawlStats.FailedCrawls)
}
