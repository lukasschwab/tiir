package web

import (
	"fmt"
	"net/http"

	"github.com/PuerkitoBio/goquery"
	"github.com/lukasschwab/tiir/pkg/text"
)

// Metadata extracts the title, author, and publication date from a given URL.
// NOTE: initial draft of this file was generated with ChatGPT. Consider the
// code experimental.
func Metadata(url string) (partial *text.Text, err error) {
	// Make HTTP GET request
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch URL: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP request failed with status: %d", resp.StatusCode)
	}

	// Parse the HTML document
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %v", err)
	}

	title := doc.Find("title").Text()

	author := doc.Find(`meta[name="author"]`).AttrOr("content", "")
	if author == "" {
		author = doc.Find(`meta[property="article:author"]`).AttrOr("content", "")
	}

	return &text.Text{Title: title, Author: author, URL: url}, nil
}
