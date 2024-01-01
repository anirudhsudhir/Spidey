package crawl

import (
	"io"
	"net/http"
	"regexp"
	"strings"
	"sync"
)

type links struct {
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

var (
	totalLinkStore   links = links{urls: make(map[string]bool)}
	totalCrawlStatus totalLinkCrawlStatus
)

func CrawlLinks(urls []string) CrawlStats {
	var wgMain sync.WaitGroup
	wgMain.Add(1)

	pingWebsites(urls, &wgMain)
	wgMain.Wait()

	totalCrawlStats := CrawlStats{}

	totalCrawlStatus.mu.Lock()
	for _, crawlStatus := range totalCrawlStatus.totalStatus {
		totalCrawlStats.TotalCrawls++
		if crawlStatus.bool {
			totalCrawlStats.SuccessfulCrawls++
		} else {
			totalCrawlStats.FailedCrawls++
		}
	}
	totalCrawlStatus.mu.Unlock()

	return totalCrawlStats
}

func pingWebsites(urls []string, wgParent *sync.WaitGroup) {
	crawlStatusChannel := make(chan crawlStatusType)
	gorountinesCreated := 0

	totalLinkStore.mu.Lock()
	for _, url := range urls {
		if totalLinkStore.urls[url] == false {

			currentUrl := strings.Trim(url, "\"")
			go fetchLinks(currentUrl, &crawlStatusChannel)
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

	wgParent.Done()
}

func fetchLinks(url string, crawlStatusChannel *chan crawlStatusType) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		crawlStatus := crawlStatusType{false, url}
		*crawlStatusChannel <- crawlStatus
		return
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		crawlStatus := crawlStatusType{false, url}
		*crawlStatusChannel <- crawlStatus
		return
	}
	defer res.Body.Close()

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		crawlStatus := crawlStatusType{false, url}
		*crawlStatusChannel <- crawlStatus
		return
	}

	linkExtractSet := regexp.MustCompile(`("http)(.*?)(")`)
	extractedLinks := linkExtractSet.FindAllString(string(resBody), -1)

	var urlSet []string
	for _, link := range extractedLinks {
		urlSet = append(urlSet, link)
	}

	if len(urlSet) > 0 {
		var wgParent sync.WaitGroup
		wgParent.Add(1)
		go pingWebsites(urlSet, &wgParent)
		wgParent.Wait()
	}

	crawlStatus := crawlStatusType{true, url}
	*crawlStatusChannel <- crawlStatus
}
