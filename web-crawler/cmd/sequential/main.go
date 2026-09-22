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

		for range levelLen {
			front := q.Front()
			q.Remove(front)
			totalPages += 1
			value := front.Value.(string)
			fmt.Printf("Crawling: %s\n", value)

			time.Sleep(delay)
			doc, err := crawler.GetURL(client, value)
			if err != nil {
				fmt.Fprintf(os.Stderr, "%v\n", err)
				continue
			}
			links, err := crawler.ExtractLinks(doc, value, allowedHost)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to extract links: %v\n", err)
				continue
			}
			for _, link := range links {
				if _, exists := seen[link]; !exists {
					seen[link] = struct{}{}
					q.PushBack(link)
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
	// if totalPages >= maxPages {
	// 	fmt.Println("Reached maximum pages limit")
	// }
}
