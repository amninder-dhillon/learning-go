package main

import (
	"container/list"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
	"web-crawler/crawler"

	"github.com/joho/godotenv"
)

type CrawlResult struct {
	URL   string
	Links []string
	Err   error
}

func worker(id int, jobs <-chan string, results chan<- CrawlResult, client *http.Client, allowedHost string, ticker *time.Ticker, wg *sync.WaitGroup, ctx context.Context) {
	defer wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		case currentURL, ok := <-jobs:
			if !ok {
				return
			}
			select {
			case <-ctx.Done():

				return
			case <-ticker.C:
				fmt.Printf("Worker %d crawling: %s\n", id, currentURL)
				result := processURL(currentURL, client, allowedHost, ctx)
				select {
				case <-ctx.Done():

					return
				case results <- result:

				}

			}
		}
	}

}

func processURL(currentURL string, client *http.Client, allowedHost string, ctx context.Context) CrawlResult {

	doc, err := crawler.GetURL(client, currentURL, ctx)
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
		log.Fatalf("loading ../../.env file: %v", err)
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
	if maxLevels <= 0 {
		log.Fatal("MAX_LEVELS must be greater than zero")
	}
	if delay <= 0 {
		log.Fatal("DELAY must be greater than zero")
	}
	if strings.TrimSpace(startURL) == "" {
		log.Fatal("START_URL must not be empty")
	}
	if strings.TrimSpace(allowedHost) == "" {
		log.Fatal("ALLOWED_HOST must not be empty")
	}
	if httpClientTimeout <= 0 {
		log.Fatal("HTTP_CLIENT_TIMEOUT must be greater than zero")
	}

	ticker := time.NewTicker(time.Duration(delay) * time.Millisecond)
	defer ticker.Stop()

	client := &http.Client{
		Timeout: time.Duration(httpClientTimeout) * time.Second,
	}
	currentLevel := 0
	totalPages := 0
	queue := list.New()

	const numWorkers int = 5
	startTime := time.Now()
	queue.PushBack(startURL)
	seen[startURL] = struct{}{}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create a channel to receive OS signals.
	// A buffer size of 1 so the notifier doesn't block.
	sigChan := make(chan os.Signal, 1)

	// Notify sigChan when an interrupt (Ctrl+C) or termination signal is received
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(sigChan)
	go func() {
		<-sigChan
		fmt.Println("\nReceived shutdown signal")
		cancel()
	}()
	for queue.Len() > 0 && currentLevel < maxLevels {
		levelLen := queue.Len()
		jobs := make(chan string)
		results := make(chan CrawlResult)
		var wg sync.WaitGroup

		workerCount := numWorkers
		if levelLen < workerCount {
			workerCount = levelLen
		}
		for w := 0; w < workerCount; w++ {
			wg.Add(1)
			go worker(w, jobs, results, client, allowedHost, ticker, &wg, ctx)
		}

		levelURLs := make([]string, 0, levelLen)
		for range levelLen {
			front := queue.Front()
			queue.Remove(front)
			totalPages += 1
			levelURLs = append(levelURLs, front.Value.(string))
		}
		go func(urls []string) {
			defer close(jobs)
			for _, pageURL := range urls {
				job := pageURL
				select {
				case <-ctx.Done():

					return
				case jobs <- job:
				}
			}
		}(levelURLs)
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
		if ctx.Err() != nil {
			break
		}
		currentLevel++
		fmt.Printf("Finished Crawling Level: %d\n", currentLevel)

	}
	if currentLevel >= maxLevels {
		fmt.Println("Reached maximum levels")
	}
	fmt.Printf("Total pages crawled by concurrent implementation: %d\n", totalPages)
	fmt.Printf("Total Duration: %v\n", time.Since(startTime))
}
