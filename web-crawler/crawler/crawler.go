package crawler

import (
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"golang.org/x/net/html"
)

var re = regexp.MustCompile(`^/wiki/([A-Za-z _]+):(.+)$`) //To match wikipedia namespaces
func NormalizeURL(u *url.URL) string {
	u.Fragment = ""
	u.RawQuery = ""
	return strings.TrimSuffix(u.String(), "/")
}

func isValidResolvedURL(resolvedURL *url.URL, allowedHost string) bool {
	if resolvedURL.Scheme != "http" && resolvedURL.Scheme != "https" {
		return false
	}
	if resolvedURL.Host != allowedHost {

		return false
	}
	if !strings.HasPrefix(resolvedURL.Path, "/wiki/") {
		return false
	}

	if re.MatchString(resolvedURL.Path) {
		return false
	}
	return true
}

func ExtractLinks(respBody *html.Node, currURL string, allowedHost string) ([]string, error) {

	baseURL, err := url.Parse(currURL)
	var links []string

	if err != nil {
		return []string{}, fmt.Errorf("parsing URL: %w", err)
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
					if isValidResolvedURL(resolvedURL, allowedHost) {

						link := NormalizeURL(resolvedURL)
						links = append(links, link)
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
	return links, nil

}

func GetURL(client *http.Client, curr string) (*html.Node, error) {
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
