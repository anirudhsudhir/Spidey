package crawl_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/anirudhsudhir/pingpong/crawl"
)

func TestPingWebsites(t *testing.T) {
	urlSetSize := 10
	var seedServers []*httptest.Server
	var testUrls []string
	var allServers []*httptest.Server
	var secondaryUrls string
	var tempServer *httptest.Server

	for i := 0; i < urlSetSize; i++ {
		secondaryUrls = "random text"
		for j := 0; j < urlSetSize; j++ {
			tempServer = createServer(50*time.Millisecond, "random text")
			secondaryUrls += "\"" + tempServer.URL + "\""
			allServers = append(allServers, tempServer)
		}
		seedServers = append(seedServers, createServer(50*time.Millisecond, secondaryUrls))
		allServers = append(allServers, seedServers[i])
		testUrls = append(testUrls, "\""+seedServers[i].URL+"\"")
	}
	defer func() {
		for i := 0; i < len(allServers); i++ {
			allServers[i].Close()
		}
	}()

	t.Run("all links crawled", func(t *testing.T) {
		got := crawl.CrawlLinks(testUrls).SuccessfulCrawls
		want := urlSetSize*urlSetSize + urlSetSize

		if got != want {
			t.Errorf("crawled %d links, want %d links", got, want)
		}
	})
}

func createServer(delay time.Duration, message string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(delay)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(message))
	}))
}
