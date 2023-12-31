package crawl

import "sync"

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
	totalLinksList.mu.Lock()
	for _, url := range urls {
		totalLinksList.urls[url] = false
	}
	totalLinksList.mu.Unlock()

	pingWebsites(urls)

	totalLinksCrawled.mu.Lock()
	linksCrawledCount := totalLinksCrawled.linksCount
	totalLinksCrawled.mu.Unlock()

	return linksCrawledCount
}

func pingWebsites(urls []string) {
	totalLinksCrawled.mu.Lock()
	totalLinksCrawled.linksCount += 5
	totalLinksCrawled.mu.Unlock()
}
