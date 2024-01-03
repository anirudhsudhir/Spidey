package crawl

import (
	"encoding/csv"
	"errors"
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
	TotalCrawls         int
	SuccessfulCrawls    int
	FailedCrawls        int
	RequestTimeExceeded int
}

type crawlStatusType struct {
	crawlStatus string
	url         string
}

type totalLinkCrawlStatus struct {
	totalStatus []crawlStatusType
	mu          sync.Mutex
}

type errorLogs struct {
	errorLog [][]byte
	mu       sync.Mutex
}

const (
	crawlSuccess             = "success"
	crawFail                 = "failure"
	crawlRequestTimeExceeded = "request time exceeded"
)

var (
	maxRequestTime          time.Duration
	maxRequestTimeInSeconds int64
	crawlerTimer            timer
	totalLinkStore          linkStore = linkStore{urls: make(map[string]bool)}
	totalCrawlStatus        totalLinkCrawlStatus
	totalErrorLogs          errorLogs
)

func CrawlLinks(urls []string, totalAllowedCrawlTime, maxAllowedRequestTime time.Duration) CrawlStats {
	maxRequestTime = maxAllowedRequestTime
	maxRequestTimeInSeconds = int64(maxRequestTime / time.Millisecond)
	statusChannel := make(chan crawlStatusType)

	go pingWebsites(urls, statusChannel)

	select {
	case <-statusChannel:
		break
	case <-time.After(totalAllowedCrawlTime):
		crawlerTimer.rmu.Lock()
		crawlerTimer.timeElapsed = true
		crawlerTimer.rmu.Unlock()
		<-statusChannel
	}

	totalCrawlStats := CrawlStats{}

	var finalCrawlData [][]string

	totalCrawlStatus.mu.Lock()
	for _, currentCrawlStatus := range totalCrawlStatus.totalStatus {
		tempStatus := make([]string, 0)
		totalCrawlStats.TotalCrawls++
		if currentCrawlStatus.crawlStatus == crawlSuccess {
			tempStatus = append(tempStatus, currentCrawlStatus.url, "crawl successful")
			totalCrawlStats.SuccessfulCrawls++
		} else if currentCrawlStatus.crawlStatus == crawlRequestTimeExceeded {
			tempStatus = append(tempStatus, currentCrawlStatus.url, "request time exceeded")
			totalCrawlStats.RequestTimeExceeded++
		} else if currentCrawlStatus.crawlStatus == crawFail {
			tempStatus = append(tempStatus, currentCrawlStatus.url, "crawl failed")
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
		if !totalLinkStore.urls[url] {

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
		currentStatus := crawlStatusType{rountineStatus.crawlStatus, rountineStatus.url}
		totalCrawlStatus.totalStatus = append(totalCrawlStatus.totalStatus, currentStatus)
		totalCrawlStatus.mu.Unlock()
	}
	completedCrawl <- crawlStatusType{}
}

func fetchLinks(url string, crawlStatusChannel chan crawlStatusType) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		recordErrorLogs(err)
		crawlStatus := crawlStatusType{crawFail, url}
		crawlStatusChannel <- crawlStatus
		return
	}

	var res *http.Response
	performRequest := func() chan struct{} {
		ch := make(chan struct{})
		go func() {
			res, err = http.DefaultClient.Do(req)
			close(ch)
		}()
		return ch
	}
	crawlerTimer.rmu.RLock()
	if crawlerTimer.timeElapsed {
		errMessage := fmt.Sprintf("request to %q cancelled as allowed time exceeded", url)
		recordErrorLogs(errors.New(errMessage))
		crawlStatus := crawlStatusType{crawlRequestTimeExceeded, url}
		crawlStatusChannel <- crawlStatus
		crawlerTimer.rmu.RUnlock()
		return
	}
	crawlerTimer.rmu.RUnlock()

	time.Sleep(2 * time.Second)

	crawlerTimer.rmu.RLock()
	if crawlerTimer.timeElapsed {
		errMessage := fmt.Sprintf("request to %q cancelled as allowed time exceeded", url)
		recordErrorLogs(errors.New(errMessage))
		crawlStatus := crawlStatusType{crawlRequestTimeExceeded, url}
		crawlStatusChannel <- crawlStatus
		crawlerTimer.rmu.RUnlock()
		return
	}
	crawlerTimer.rmu.RUnlock()

	select {
	case <-performRequest():
		if err != nil {
			recordErrorLogs(err)
			crawlStatus := crawlStatusType{crawFail, url}
			crawlStatusChannel <- crawlStatus
			return
		}
	case <-time.After(maxRequestTime):
		errMessage := fmt.Sprintf("%q took more than %d milliseconds to respond", url, maxRequestTimeInSeconds)
		recordErrorLogs(errors.New(errMessage))
		crawlStatus := crawlStatusType{crawlRequestTimeExceeded, url}
		crawlStatusChannel <- crawlStatus
		return
	}
	defer res.Body.Close()

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		recordErrorLogs(err)
		crawlStatus := crawlStatusType{crawlSuccess, url}
		crawlStatusChannel <- crawlStatus
		return
	}

	linkExtractSet := regexp.MustCompile(`("http)(.*?)(")`)
	extractedLinks := linkExtractSet.FindAllString(string(resBody), -1)

	var urlSet []string
	urlSet = append(urlSet, extractedLinks...)

	if len(urlSet) > 0 {
		statusChannel := make(chan crawlStatusType)
		go pingWebsites(urlSet, statusChannel)
		<-statusChannel
	}

	crawlStatus := crawlStatusType{crawlSuccess, url}
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
