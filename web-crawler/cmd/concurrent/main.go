package main

import (
	"container/list"
	"fmt"
	"net/http"
	"sync"
	"time"
	"web-crawler/crawler"
)

type CrawlResult struct {
	URL   string
	Links []string
	Err   error
}

func main() {
	seen := make(map[string]struct{})
	//maxPages := 100
	maxLevels := 2
	currentLevel := 0
	totalPages := 0
	const delay = 250 * time.Millisecond
	startURL := "https://en.wikipedia.org/wiki/Miss_Meyers"
	allowedHost := "en.wikipedia.org"
	client := &http.Client{
		Timeout: 5 * time.Second,
	}
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
				links, _ := crawler.ExtractLinks(doc, currURL, allowedHost)
				results <- CrawlResult{
					URL:   currURL,
					Err:   nil,
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
