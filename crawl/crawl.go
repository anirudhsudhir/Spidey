package crawl

import (
	"io"
	"net/http"
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

type linkResponses struct {
	linkResponse string
	mu           sync.Mutex
}

var (
	totalLinksList     links         = links{urls: make(map[string]bool)}
	totalLinksCrawled  linksCrawled  = linksCrawled{}
	totalLinkResponses linkResponses = linkResponses{}
)

func CrawlLinks(urls []string) int {
	pingWebsites(urls)

	totalLinksCrawled.mu.Lock()
	linksCrawledCount := totalLinksCrawled.linksCount
	totalLinksCrawled.mu.Unlock()

	return linksCrawledCount
}

func GetResponses() string {
	totalLinkResponses.mu.Lock()
	responses := totalLinkResponses.linkResponse
	totalLinkResponses.mu.Unlock()

	return responses
}

func pingWebsites(urls []string) {
	var wg sync.WaitGroup
	wg.Add(len(urls))

	for _, url := range urls {
		totalLinksList.mu.Lock()

		if totalLinksList.urls[url] == false {
			totalLinksList.urls[url] = true

			totalLinksCrawled.mu.Lock()
			totalLinksCrawled.linksCount++
			totalLinksCrawled.mu.Unlock()

			currentUrl := url
			go fetchLinks(currentUrl, &wg)

		}
		totalLinksList.mu.Unlock()
	}
	wg.Wait()
}

func fetchLinks(url string, wg *sync.WaitGroup) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		totalLinksCrawled.mu.Lock()
		totalLinksCrawled.failedRequests++
		totalLinksCrawled.mu.Unlock()

		return
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		totalLinksCrawled.mu.Lock()
		totalLinksCrawled.failedRequests++
		totalLinksCrawled.mu.Unlock()

		return
	}
	defer res.Body.Close()

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		totalLinksCrawled.mu.Lock()
		totalLinksCrawled.failedRequests++
		totalLinksCrawled.mu.Unlock()

		return
	}

	totalLinkResponses.mu.Lock()
	totalLinkResponses.linkResponse += string(resBody)
	totalLinkResponses.mu.Unlock()

	wg.Done()
}
