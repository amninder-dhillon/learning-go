package main

import (
	"container/list"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
	"web-crawler/crawler"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load("../../.env")
	if err != nil {
		log.Fatal("Error loading ../../.env file")
	}

	seen := make(map[string]struct{})

	// maxPages := 100
	maxLevels, err := strconv.Atoi(os.Getenv("MAX_LEVELS"))
	if err != nil {
		log.Fatalf("invalid MAX_LEVELS: %v", err)
	}

	delayMilliseconds, err := strconv.Atoi(os.Getenv("DELAY"))
	if err != nil {
		log.Fatalf("invalid DELAY: %v", err)
	}
	delay := time.Duration(delayMilliseconds) * time.Millisecond

	startURL := os.Getenv("START_URL")
	allowedHost := os.Getenv("ALLOWED_HOST")
	httpClientTimeout, err := strconv.Atoi(os.Getenv("HTTP_CLIENT_TIMEOUT"))
	if err != nil {
		log.Fatalf("invalid HTTP_CLIENT_TIMEOUT: %v", err)
	}
	client := &http.Client{
		Timeout: time.Duration(httpClientTimeout) * time.Second,
	}
	currentLevel := 0
	totalPages := 0
	queue := list.New()
	startTime := time.Now()
	queue.PushBack(startURL)
	seen[startURL] = struct{}{}
	for queue.Len() > 0 && currentLevel < maxLevels {
		levelLen := queue.Len()

		for range levelLen {
			front := queue.Front()
			queue.Remove(front)
			totalPages += 1
			pageURL := front.Value.(string)
			fmt.Printf("Crawling: %s\n", pageURL)

			time.Sleep(delay)
			doc, err := crawler.GetURL(client, pageURL)
			if err != nil {
				fmt.Fprintf(os.Stderr, "%v\n", err)
				continue
			}
			links, err := crawler.ExtractLinks(doc, pageURL, allowedHost)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to extract links: %v\n", err)
				continue
			}
			for _, link := range links {
				if _, exists := seen[link]; !exists {
					seen[link] = struct{}{}
					queue.PushBack(link)
				}
			}
		}
		fmt.Printf("Finished Crawling Level: %d\n", currentLevel)
		currentLevel += 1

	}
	if currentLevel >= maxLevels {
		fmt.Println("Reached maximum levels")
	}
	fmt.Printf("Total pages crawled by sequential implementation: %d\n", totalPages)
	fmt.Printf("Total Duration: %v\n", time.Since(startTime))
	// if totalPages >= maxPages {
	// 	fmt.Println("Reached maximum pages limit")
	// }
}
