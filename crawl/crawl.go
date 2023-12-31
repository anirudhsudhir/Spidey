package crawl

import (
	"io"
	"net/http"
	"regexp"
	"sync"
)

type links struct {
	urls map[string]bool
	mu   sync.Mutex
}

type linksCrawled struct {
	linksCount     int
	failedRequests int
	mu             sync.Mutex
}

var (
	totalLinksList    links        = links{urls: make(map[string]bool)}
	totalLinksCrawled linksCrawled = linksCrawled{}
)

func CrawlLinks(urls []string) int {
	var wgMain sync.WaitGroup
	wgMain.Add(1)

	pingWebsites(urls, &wgMain)
	wgMain.Wait()

	totalLinksCrawled.mu.Lock()
	linksCrawledCount := totalLinksCrawled.linksCount
	totalLinksCrawled.mu.Unlock()

	return linksCrawledCount
}

func pingWebsites(urls []string, wgParent *sync.WaitGroup) {
	var wg sync.WaitGroup

	for _, url := range urls {
		totalLinksList.mu.Lock()

		if totalLinksList.urls[url] == false {
			totalLinksList.urls[url] = true

			totalLinksCrawled.mu.Lock()
			totalLinksCrawled.linksCount++
			totalLinksCrawled.mu.Unlock()

			currentUrl := url
			wg.Add(1)
			go fetchLinks(currentUrl, &wg)

		}
		totalLinksList.mu.Unlock()
	}
	wg.Wait()
	wgParent.Done()
}

func fetchLinks(url string, wg *sync.WaitGroup) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		totalLinksCrawled.mu.Lock()
		totalLinksCrawled.failedRequests++
		totalLinksCrawled.mu.Unlock()

		wg.Done()
		return
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		totalLinksCrawled.mu.Lock()
		totalLinksCrawled.failedRequests++
		totalLinksCrawled.mu.Unlock()

		wg.Done()
		return
	}
	defer res.Body.Close()

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		totalLinksCrawled.mu.Lock()
		totalLinksCrawled.failedRequests++
		totalLinksCrawled.mu.Unlock()

		wg.Done()
		return
	}

	linkExtractSet := regexp.MustCompile(`(http)(.*?)( )`)
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

	wg.Done()
}
