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

func main() {
	err := godotenv.Load("../../.env")
	if err != nil {
		log.Fatal("Error loading ../../.env file")
	}

	seen := make(map[string]struct{})

	//maxPages := 100
	maxLevels, _ := strconv.Atoi(os.Getenv("MAX_LEVELS"))
	env_delay, _ := strconv.Atoi(os.Getenv("DELAY"))
	delay := time.Duration(env_delay) * time.Millisecond
	startURL := os.Getenv("START_URL")
	allowedHost := os.Getenv("ALLOWED_HOST")
	http_client_timeout, _ := strconv.Atoi(os.Getenv("HTTP_CLIENT_TIMEOUT"))
	client := &http.Client{
		Timeout: time.Duration(http_client_timeout) * time.Second,
	}
	currentLevel := 0
	totalPages := 0
	q := list.New()

	q.PushBack(startURL)
	seen[startURL] = struct{}{}
	for q.Len() > 0 && currentLevel < maxLevels {
		levelLen := q.Len()
		results := make(chan CrawlResult, levelLen)
		var wg sync.WaitGroup

		for range levelLen {
			front := q.Front()
			q.Remove(front)
			totalPages += 1
			value := front.Value.(string)
			wg.Add(1)
			go func(currURL string) {
				defer wg.Done()

				fmt.Printf("Crawling: %s\n", currURL)

				time.Sleep(delay)
				doc, err := crawler.GetURL(client, currURL)
				if err != nil {
					results <- CrawlResult{
						Err: err,
						URL: currURL,
					}
					return
				}
				links, err := crawler.ExtractLinks(doc, currURL, allowedHost)
				results <- CrawlResult{
					URL:   currURL,
					Err:   err,
					Links: links,
				}
			}(value)
		}
		go func() {
			wg.Wait()
			close(results)
		}()
		currentLevel += 1
		for result := range results {

			for _, link := range result.Links {
				if _, exists := seen[link]; exists {
					continue
				}
				seen[link] = struct{}{}
				q.PushBack(link)
			}
		}
		fmt.Printf("Finished Crawling Level: %d\n", currentLevel)
	}
	if currentLevel >= maxLevels {
		fmt.Println("Reached maximum levels")
	}
	fmt.Printf("Total pages crawled by concurrent implementation: %d\n", totalPages)

}
