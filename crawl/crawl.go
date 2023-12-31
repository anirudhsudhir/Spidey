package crawl

import (
	"net/http"
	"sync"
)

type links struct {
	urls map[string]bool
	mu   sync.Mutex
}

type linksCrawled struct {
	linksCount int
	mu         sync.Mutex
}

var (
	totalLinksList    links        = links{urls: make(map[string]bool)}
	totalLinksCrawled linksCrawled = linksCrawled{}
)

func CrawlLinks(urls []string) int {
	pingWebsites(urls)

	totalLinksCrawled.mu.Lock()
	linksCrawledCount := totalLinksCrawled.linksCount
	totalLinksCrawled.mu.Unlock()

	return linksCrawledCount
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
	http.Get(url)
	wg.Done()
}
