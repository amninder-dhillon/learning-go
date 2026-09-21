package main

import (
	"container/list"
	"fmt"
	"net/http"
	"os"
	"time"
	"web-crawler/crawler"
)

func main() {
	seen := make(map[string]struct{})
	//maxPages := 100
	maxLevels := 2
	currentLevel := 0
	//totalPages := 0
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

		for range levelLen {
			front := q.Front()
			q.Remove(front)
			//totalPages += 1
			value := front.Value.(string)
			fmt.Printf("Crawling: %s\n", value)

			time.Sleep(delay)
			doc, err := crawler.GetURL(client, value)
			if err != nil {
				fmt.Fprintf(os.Stderr, "%v\n", err)
				continue
			}
			if err := crawler.ExtractLinks(doc, value, q, seen, allowedHost); err != nil {
				fmt.Fprintf(os.Stderr, "Failed to extract links: %v\n", err)
				continue
			}
		}
		fmt.Printf("Finished Crawling Level: %d\n", currentLevel)
		currentLevel += 1

	}
	if currentLevel >= maxLevels {
		fmt.Println("Reached maximum levels")
	}
	// if totalPages >= maxPages {
	// 	fmt.Println("Reached maximum pages limit")
	// }
}
