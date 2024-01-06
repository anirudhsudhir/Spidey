package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"time"

	"github.com/anirudhsudhir/spidey/crawl"
)

func main() {
	totalCrawlTime, maxRequestTime := parseArguments()
	seedUrls := parseSeeds()
	crawlStats := crawl.CrawlLinks(seedUrls, totalCrawlTime, maxRequestTime)
	fmt.Printf("TotalCrawls: %d, SuccessfulCrawls: %d, FailedCrawls: %d, Request Time Exceeded: %d", crawlStats.TotalCrawls, crawlStats.SuccessfulCrawls, crawlStats.FailedCrawls, crawlStats.RequestTimeExceeded)
}

func parseSeeds() (seedUrls []string) {
	seedFile, err := os.Open("seeds.txt")
	if err != nil {
		fmt.Printf("Error while opening seeds.txt: %q\n", err)
		os.Exit(1)
	}
	defer seedFile.Close()
	scanner := bufio.NewScanner(seedFile)
	for scanner.Scan() {
		re := regexp.MustCompile(`("http)(.*?)(")`)
		match := re.FindString(scanner.Text())
		if match != "" {
			seedUrls = append(seedUrls, match)
		}
	}
	err = scanner.Err()
	if err != nil {
		fmt.Printf("Error reading seeds.txt: %q\n", err)
		os.Exit(1)
	}
	return
}

func parseArguments() (time.Duration, time.Duration) {
	if len(os.Args) < 3 {
		fmt.Println("Invalid number of arguments")
		os.Exit(1)
	}
	crawlTime, err := strconv.Atoi(os.Args[1])
	if err != nil {
		fmt.Println("Error while reading arguments: ", err)
	}
	requestTime, err := strconv.Atoi(os.Args[2])
	if err != nil {
		fmt.Println("Error while reading arguments: ", err)
	}
	totalCrawlTime := time.Duration(crawlTime) * time.Millisecond
	maxRequestTime := time.Duration(requestTime) * time.Millisecond
	return totalCrawlTime, maxRequestTime
}
