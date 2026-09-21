package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

func getUrls() ([]string, error) {
	data, err := os.ReadFile("urls.json")
	if err != nil {
		return nil, fmt.Errorf("reading urls.json: %w", err)
	}

	var urls []string
	if err := json.Unmarshal(data, &urls); err != nil {
		return nil, fmt.Errorf("decoding urls.json: %w", err)
	}

	return urls, nil
}

type Result struct {
	URL        string
	StatusCode int
	Duration   time.Duration
	Err        error
	Index      int
}

func checkUrl(client http.Client, url string, index int) Result {

	start := time.Now()

	resp, err := client.Get(url)

	duration := time.Since(start)
	if err != nil {
		return Result{Duration: duration, Err: err, URL: url, Index: index}
	}
	defer resp.Body.Close()

	return Result{StatusCode: resp.StatusCode, Duration: duration, Err: nil, URL: url, Index: index}
}

func printResult(result Result) {
	if result.Err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", result.Err)
		return
	}

	fmt.Printf("Index = %d, Status Code = %d and Duration = %v from %s\n", result.Index, result.StatusCode, result.Duration, result.URL)
	fmt.Println("=========================================")

}

func sequentialUrlCheck(client http.Client, urls []string) {
	for i, url := range urls {
		result := checkUrl(client, url, i)
		printResult(result)
	}

}

func concurrentUrlCheck(client http.Client, urls []string) {
	results := make(chan Result, len(urls))

	for i, url := range urls {

		// We cannot do:
		//
		//     go checkURL(client, url, index)
		//
		// if we need to use the Result returned by checkURL. A goroutine runs
		// asynchronously, so its return value cannot be assigned directly:
		//
		//     result := go checkURL(...) // invalid
		//
		// Instead, the anonymous function acts as an adapter: it calls checkURL,
		// receives the returned Result, and sends that Result through the channel:
		//
		//     go func(...) {
		//         results <- checkURL(...)
		//     }(...)
		//
		// Another valid design would be to pass the results channel directly into
		// checkURL. Then checkURL could send its result itself, allowing:
		//
		//     go checkURL(client, url, index, results)
		//
		// If the return value were not needed at all, `go checkURL(...)` would be
		// perfectly valid. However, the program would still need some synchronization
		// mechanism (such as a channel or WaitGroup) to keep main alive until the
		// goroutines finish. When main returns, the entire program exits, including
		// any goroutines that are still running.
		go func(url string, index int) {
			results <- checkUrl(client, url, index)
		}(url, i)
	}

	for range len(urls) {

		result := <-results
		printResult(result)
	}
}

func main() {
	urls, err := getUrls()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading URLs: %v\n", err)
		return
	}
	client := &http.Client{
		Timeout: 5 * time.Second, //Timeout after 5 seconds
	}
	sequentialStartTime := time.Now()
	sequentialUrlCheck(*client, urls)

	fmt.Printf("Sequential URL checking is complete in %v\n", time.Since(sequentialStartTime))

	concurrentStartTime := time.Now()
	concurrentUrlCheck(*client, urls)
	fmt.Printf("Concurrent URL checking is complete in %v\n", time.Since(concurrentStartTime))

}
