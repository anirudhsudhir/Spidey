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

type CrawlStats struct {
	TotalCrawls      int
	SuccessfulCrawls int
	FailedCrawls     int
}

var totalLinkStore links = links{urls: make(map[string]bool)}

func CrawlLinks(urls []string) CrawlStats {
	var wgMain sync.WaitGroup
	wgMain.Add(1)

	pingWebsites(urls, &wgMain)
	wgMain.Wait()

	totalCrawlStats := CrawlStats{}

	totalLinkStore.mu.Lock()
	for _, crawlStatus := range totalLinkStore.urls {
		totalCrawlStats.TotalCrawls++
		if crawlStatus {
			totalCrawlStats.SuccessfulCrawls++
		} else {
			totalCrawlStats.FailedCrawls++
		}
	}
	totalLinkStore.mu.Unlock()

	return totalCrawlStats
}

func pingWebsites(urls []string, wgParent *sync.WaitGroup) {
	var wg sync.WaitGroup

	totalLinkStore.mu.Lock()
	for _, url := range urls {
		if totalLinkStore.urls[url] == false {
			totalLinkStore.urls[url] = true

			currentUrl := url
			wg.Add(1)
			go fetchLinks(currentUrl, &wg)

		}
	}
	totalLinkStore.mu.Unlock()
	wg.Wait()
	wgParent.Done()
}

func fetchLinks(url string, wg *sync.WaitGroup) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		// Return stats
		wg.Done()
		return
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		// Return stats

		wg.Done()
		return
	}
	defer res.Body.Close()

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		// Return stats

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
