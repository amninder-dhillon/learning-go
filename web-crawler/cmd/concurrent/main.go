package main

import (
	"container/list"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"
	"web-crawler/crawler"

	"github.com/joho/godotenv"
)

type CrawlResult struct {
	URL   string
	Links []string
	Err   error
}

func worker(id int, jobs <-chan string, results chan<- CrawlResult, client *http.Client, allowedHost string, ticker *time.Ticker, wg *sync.WaitGroup) {
	defer wg.Done()
	for currentURL := range jobs {
		<-ticker.C
		fmt.Printf("Worker %d crawling: %s\n", id, currentURL)
		results <- processURL(currentURL, client, allowedHost)
	}
}

func processURL(currentURL string, client *http.Client, allowedHost string) CrawlResult {

	doc, err := crawler.GetURL(client, currentURL)
	if err != nil {
		return CrawlResult{
			Err: err,
			URL: currentURL,
		}

	}
	links, err := crawler.ExtractLinks(doc, currentURL, allowedHost)
	return CrawlResult{
		URL:   currentURL,
		Err:   err,
		Links: links,
	}
}

func main() {
	err := godotenv.Load("../../.env")
	if err != nil {
		log.Fatal("Error loading ../../.env file")
	}

	seen := make(map[string]struct{})

	//maxPages := 100

	maxLevels, err := strconv.Atoi(os.Getenv("MAX_LEVELS"))
	if err != nil {
		log.Fatalf("invalid MAX_LEVELS: %v", err)
	}

	delay, err := strconv.Atoi(os.Getenv("DELAY"))
	if err != nil {
		log.Fatalf("invalid DELAY: %v", err)
	}

	startURL := os.Getenv("START_URL")
	allowedHost := os.Getenv("ALLOWED_HOST")
	httpClientTimeout, err := strconv.Atoi(os.Getenv("HTTP_CLIENT_TIMEOUT"))
	if err != nil {
		log.Fatalf("invalid HTTP_CLIENT_TIMEOUT: %v", err)
	}

	ticker := time.NewTicker(time.Duration(delay) * time.Millisecond)

	client := &http.Client{
		Timeout: time.Duration(httpClientTimeout) * time.Second,
	}
	currentLevel := 0
	totalPages := 0
	queue := list.New()

	const numWorkers int = 5

	queue.PushBack(startURL)
	seen[startURL] = struct{}{}
	for queue.Len() > 0 && currentLevel < maxLevels {
		levelLen := queue.Len()
		results := make(chan CrawlResult, levelLen)
		jobs := make(chan string)
		var wg sync.WaitGroup
		for w := range numWorkers {
			wg.Add(1)
			go worker(w, jobs, results, client, allowedHost, ticker, &wg)
		}
		for range levelLen {
			front := queue.Front()
			queue.Remove(front)
			totalPages += 1
			pageURL := front.Value.(string)
			jobs <- pageURL

		}
		close(jobs)
		go func() {
			wg.Wait()
			close(results)
		}()

		for result := range results {
			if result.Err != nil {
				fmt.Fprintf(
					os.Stderr,
					"failed crawling %s: %v\n",
					result.URL,
					result.Err,
				)
				continue
			}

			for _, link := range result.Links {
				if _, exists := seen[link]; exists {
					continue
				}
				seen[link] = struct{}{}
				queue.PushBack(link)
			}
		}

		currentLevel++
		fmt.Printf("Finished Crawling Level: %d\n", currentLevel)
	}
	ticker.Stop()
	if currentLevel >= maxLevels {
		fmt.Println("Reached maximum levels")
	}
	fmt.Printf("Total pages crawled by concurrent implementation: %d\n", totalPages)

}
