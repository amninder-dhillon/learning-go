package main

import (
	"container/list"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	"golang.org/x/net/html"
)

func extractLinks(respBody *html.Node, currUrl string, q *list.List, visited map[string]struct{}) error {

	baseURL, err := url.Parse(currUrl)

	if err != nil {
		return fmt.Errorf("Error parsing the URL: %v", err)
	}
	// Recursive function to traverse the HTML node tree
	var visitErr error
	var visit func(*html.Node)
	visit = func(n *html.Node) {
		if visitErr != nil {
			return
		}
		// Target only element nodes with the tag name "a"
		if n.Type == html.ElementNode && n.Data == "a" {
			for _, attr := range n.Attr {
				if attr.Key == "href" {
					hrefURL, err := url.Parse(attr.Val)
					if err != nil {
						visitErr = fmt.Errorf("error parsing the value: %v", err)
						return
					}
					resolvedURL := baseURL.ResolveReference(hrefURL)
					link := strings.TrimSuffix(resolvedURL.String(), "/")
					_, exists := visited[link]
					if !exists {
						visited[link] = struct{}{}
						q.PushBack(link)
					}
				}
			}
		}
		// Traverse down to children and across to siblings
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			visit(c)
		}
	}

	visit(respBody)
	return visitErr
}

func getURL(curr string) (*html.Node, error) {
	resp, err := http.Get(curr)
	if err != nil {
		return nil, fmt.Errorf("Error while getting the url: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Received a status code other than 200: %d", resp.StatusCode)
	}
	doc, err := html.Parse(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse html for url %s: %v", curr, err)
	}
	return doc, nil
}
func main() {
	visited := make(map[string]struct{})

	startUrl := "https://golang.org"
	q := list.New()
	q.PushBack(startUrl)
	visited[startUrl] = struct{}{}
	for q.Len() > 0 {
		front := q.Front()
		q.Remove(front)
		fmt.Printf("Crawling: %s\n", front.Value.(string))
		doc, err := getURL(front.Value.(string))
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v", err)
			continue
		}
		if err := extractLinks(doc, front.Value.(string), q, visited); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to extract links: %v", err)
			continue
		}
	}
}
