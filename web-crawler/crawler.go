package main

import (
	"container/list"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"

	"golang.org/x/net/html"
)

func normalizeURL(u *url.URL) string {
	u.Fragment = ""
	u.RawQuery = ""
	return strings.TrimSuffix(u.String(), "/")
}

func extractLinks(respBody *html.Node, currUrl string, q *list.List, seen map[string]struct{}, allowedHost string) error {

	baseURL, err := url.Parse(currUrl)
	re := regexp.MustCompile(`^/wiki/([A-Za-z _]+):(.+)$`) //To match wikipedia namespaces

	if err != nil {
		return fmt.Errorf("Error parsing the URL: %w", err)
	}
	// Recursive function to traverse the HTML node tree

	var visit func(*html.Node)
	visit = func(n *html.Node) {

		// Target only element nodes with the tag name "a"
		if n.Type == html.ElementNode && n.Data == "a" {
			for _, attr := range n.Attr {
				if attr.Key == "href" {
					hrefURL, err := url.Parse(attr.Val)

					if err != nil {
						fmt.Printf("error parsing the value: %v\n", err)
						continue
					}
					resolvedURL := baseURL.ResolveReference(hrefURL)
					if resolvedURL.Scheme != "http" && resolvedURL.Scheme != "https" {
						continue
					}
					if resolvedURL.Host != allowedHost {

						continue
					}
					if !strings.HasPrefix(resolvedURL.Path, "/wiki/") {
						continue
					}

					if re.MatchString(resolvedURL.Path) {
						continue
					}
					link := normalizeURL(resolvedURL)
					_, exists := seen[link]
					if !exists {
						seen[link] = struct{}{}
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
	return nil
}

func getURL(client *http.Client, curr string) (*html.Node, error) {
	req, err := http.NewRequest(http.MethodGet, curr, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set(
		"User-Agent",
		"GoCrawler/1.0 (your-email@example.com)",
	)

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("getting URL %s: %w", curr, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(
			"received status %s for %s",
			resp.Status,
			curr,
		)
	}
	contentType := resp.Header.Get("Content-Type")
	if !strings.Contains(contentType, "text/html") {
		return nil, fmt.Errorf("unsupported content type: %s", contentType)
	}
	doc, err := html.Parse(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("parsing HTML for %s: %w", curr, err)
	}

	return doc, nil
}
func main() {
	seen := make(map[string]struct{})
	maxPages := 100
	totalPages := 0
	const delay = 250 * time.Millisecond
	startUrl := "https://en.wikipedia.org/wiki/Punjab"
	allowedHost := "en.wikipedia.org"
	client := &http.Client{
		Timeout: 5 * time.Second,
	}
	q := list.New()

	q.PushBack(startUrl)
	seen[startUrl] = struct{}{}
	for q.Len() > 0 && totalPages < maxPages {
		front := q.Front()
		q.Remove(front)

		totalPages += 1
		value := front.Value.(string)
		fmt.Printf("Crawling: %s\n", value)

		time.Sleep(delay)
		doc, err := getURL(client, value)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			continue
		}
		if err := extractLinks(doc, value, q, seen, allowedHost); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to extract links: %v\n", err)
			continue
		}
	}
	if totalPages >= maxPages {
		fmt.Println("Reached maximum pages limit")
	}
}
