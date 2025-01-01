package web

import (
	"fmt"
	"io"
	"net/http"

	"github.com/PuerkitoBio/goquery"
	"github.com/lukasschwab/tiir/pkg/text"
)

// WebMetadata extracts the title, author, and publication date from a given URL.
// NOTE: initial draft of this file was generated with ChatGPT. Consider the
// code experimental.
func WebMetadata(url string) (partial *text.Text, err error) {
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
	text, err := Metadata(resp.Body)
	if text != nil {
		text.URL = url
	}
	return text, err
}

func Metadata(source io.Reader) (*text.Text, error) {
	doc, err := goquery.NewDocumentFromReader(source)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %v", err)
	}

	title := doc.Find("title").Text()

	author := doc.Find(`meta[name="author"]`).AttrOr("content", "")
	if author == "" {
		author = doc.Find(`meta[property="article:author"]`).AttrOr("content", "")
	}
	if author == "" {
		author = doc.Find(`meta[name="creator"]`).AttrOr("content", "")
	}

	return &text.Text{Title: title, Author: author}, nil
}
