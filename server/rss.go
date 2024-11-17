package server

import (
	"encoding/xml"
	"fmt"
	"io"
	"morbo/errors"
	"net/http"
)

type RSS struct {
	XMLName xml.Name `xml:"rss"`
	Channel Channel  `xml:"channel"`
}

type Channel struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	Items       []Item `xml:"item"`
}

type Item struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
}

func parseRSS(url string) (*RSS, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, errors.Chain(fmt.Sprintf("failed to request the RSS feed at %s", url), err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.Chain(fmt.Sprintf("didn't find the RSS feed at %s", url), err)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Chain(fmt.Sprintf("failed to read the RSS feed at %s", url), err)
	}

	var rss RSS
	err = xml.Unmarshal(body, &rss)
	if err != nil {
		return nil, errors.Chain(fmt.Sprintf("failed to parse the RSS feed at %s", url), err)
	}

	return &rss, nil
}
