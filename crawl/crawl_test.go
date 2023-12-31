package crawl_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/anirudhsudhir/pingpong/crawl"
)

func TestPingWebsites(t *testing.T) {
	totalUrls := 5
	var testServers []*httptest.Server
	var testUrls []string

	for i := 1; i <= totalUrls; i++ {
		testServers = append(testServers, createServer(time.Duration(100+i*50)*time.Millisecond, string(rune(i+64))))
		testUrls = append(testUrls, testServers[i-1].URL)
	}
	defer func() {
		for i := 0; i < totalUrls; i++ {
			testServers[i].Close()
		}
	}()

	t.Run("all links crawled", func(t *testing.T) {
		got := crawl.CrawlLinks(testUrls)
		want := totalUrls

		if got != want {
			t.Errorf("crawled %q links, want %q links", got, want)
		}
	})

	t.Run("link reponses matching", func(t *testing.T) {
		got := crawl.GetResponses()
		want := "ABCDE"

		if got != want {
			t.Errorf("got %q as body, want %q as body", got, want)
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
