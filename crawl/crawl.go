package crawl

import (
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"
)

type timer struct {
	timeElapsed bool
	rmu         sync.RWMutex
}

type linkStore struct {
	urls map[string]bool
	mu   sync.Mutex
}

type CrawlStats struct {
	TotalCrawls      int
	SuccessfulCrawls int
	FailedCrawls     int
}

type linkCrawlStatusType struct {
	url             string
	successfulCrawl bool
	mu              sync.Mutex
}

type crawlStatusType struct {
	bool
	string
}

type totalLinkCrawlStatus struct {
	totalStatus []crawlStatusType
	mu          sync.Mutex
}

type errorLogs struct {
	errorLog [][]byte
	mu       sync.Mutex
}

var (
	crawlerTimer     timer
	totalLinkStore   linkStore = linkStore{urls: make(map[string]bool)}
	totalCrawlStatus totalLinkCrawlStatus
	totalErrorLogs   errorLogs
)

func CrawlLinks(urls []string, allowedRuntime time.Duration) CrawlStats {
	statusChannel := make(chan crawlStatusType)

	go pingWebsites(urls, statusChannel)

	select {
	case <-statusChannel:
		break
	case <-time.After(allowedRuntime):
		crawlerTimer.rmu.Lock()
		crawlerTimer.timeElapsed = true
		crawlerTimer.rmu.Unlock()
		<-statusChannel
	}

	totalCrawlStats := CrawlStats{}

	var finalCrawlData [][]string

	totalCrawlStatus.mu.Lock()
	for _, crawlStatus := range totalCrawlStatus.totalStatus {
		tempStatus := make([]string, 0)
		totalCrawlStats.TotalCrawls++
		if crawlStatus.bool {
			tempStatus = append(tempStatus, crawlStatus.string, "crawl successful")
			totalCrawlStats.SuccessfulCrawls++
		} else {
			tempStatus = append(tempStatus, crawlStatus.string, "crawl failed")
			totalCrawlStats.FailedCrawls++
		}
		finalCrawlData = append(finalCrawlData, tempStatus)
	}
	totalCrawlStatus.mu.Unlock()

	writeCrawlLogs(finalCrawlData)

	totalErrorLogs.mu.Lock()
	writeErrorLogs(totalErrorLogs.errorLog)
	totalErrorLogs.mu.Unlock()

	return totalCrawlStats
}

func pingWebsites(urls []string, completedCrawl chan crawlStatusType) {
	crawlerTimer.rmu.RLock()
	if crawlerTimer.timeElapsed {
		completedCrawl <- crawlStatusType{}
		crawlerTimer.rmu.RUnlock()
		return
	}
	crawlerTimer.rmu.RUnlock()

	crawlStatusChannel := make(chan crawlStatusType)
	gorountinesCreated := 0

	totalLinkStore.mu.Lock()
	for _, url := range urls {
		if totalLinkStore.urls[url] == false {

			currentUrl := strings.Trim(url, "\"")
			go fetchLinks(currentUrl, crawlStatusChannel)
			totalLinkStore.urls[url] = true
			gorountinesCreated++

		}
	}
	totalLinkStore.mu.Unlock()

	for i := 1; i <= gorountinesCreated; i++ {
		rountineStatus := <-crawlStatusChannel
		totalCrawlStatus.mu.Lock()
		currentStatus := crawlStatusType{rountineStatus.bool, rountineStatus.string}
		totalCrawlStatus.totalStatus = append(totalCrawlStatus.totalStatus, currentStatus)
		totalCrawlStatus.mu.Unlock()
	}
	completedCrawl <- crawlStatusType{}
}

func fetchLinks(url string, crawlStatusChannel chan crawlStatusType) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		recordErrorLogs(err)
		crawlStatus := crawlStatusType{false, url}
		crawlStatusChannel <- crawlStatus
		return
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		recordErrorLogs(err)
		crawlStatus := crawlStatusType{false, url}
		crawlStatusChannel <- crawlStatus
		return
	}
	defer res.Body.Close()

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		recordErrorLogs(err)
		crawlStatus := crawlStatusType{false, url}
		crawlStatusChannel <- crawlStatus
		return
	}

	linkExtractSet := regexp.MustCompile(`("http)(.*?)(")`)
	extractedLinks := linkExtractSet.FindAllString(string(resBody), -1)

	var urlSet []string
	for _, link := range extractedLinks {
		urlSet = append(urlSet, link)
	}

	if len(urlSet) > 0 {
		statusChannel := make(chan crawlStatusType)
		go pingWebsites(urlSet, statusChannel)
		<-statusChannel
	}

	crawlStatus := crawlStatusType{true, url}
	crawlStatusChannel <- crawlStatus
}

func writeCrawlLogs(finalCrawlData [][]string) {
	file, err := os.Create("crawl_data.csv")
	if err != nil {
		recordErrorLogs(err)
		return
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	for _, crawlData := range finalCrawlData {
		err := writer.Write(crawlData)
		if err != nil {
			recordErrorLogs(err)
			return
		}
	}
	writer.Flush()
}

func writeErrorLogs(finalErrorLogs [][]byte) {
	file, err := os.Create("log.txt")
	if err != nil {
		fmt.Printf("Encountered error while create error log file: %v", err)
		return
	}
	defer file.Close()

	for _, log := range finalErrorLogs {
		_, err = file.Write(log)
		if err != nil {
			fmt.Printf("Encountered error while writing error logs to file: %v", err)
			return
		}
	}
}

func recordErrorLogs(err error) {
	totalErrorLogs.mu.Lock()
	totalErrorLogs.errorLog = append(totalErrorLogs.errorLog, []byte(err.Error()+"\n"))
	totalErrorLogs.mu.Unlock()
}
