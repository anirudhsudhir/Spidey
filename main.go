package main

import (
	"fmt"
	"time"

	"github.com/anirudhsudhir/spidey/crawl"
)

func main() {
	seedUrls := []string{}
	crawlStats := crawl.CrawlLinks(seedUrls, 1*time.Second, time.Second)
	fmt.Printf("TotalCrawls: %d, SuccessfulCrawls: %d, FailedCrawls: %d, Request Time Exceeded: %d", crawlStats.TotalCrawls, crawlStats.SuccessfulCrawls, crawlStats.FailedCrawls, crawlStats.RequestTimeExceeded)
}
