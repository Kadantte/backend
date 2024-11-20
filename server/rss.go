package server

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"

	"morbo/errors"
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

type statusCode = int

func parseRSS(url string) (*RSS, error, statusCode) {
	resp, err := http.Get(url)
	if err != nil {
		err := errors.Chain(fmt.Sprintf("failed to request the RSS feed at %s", url), err)
		return nil, err, http.StatusBadRequest
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err := fmt.Errorf("got a non-ok response code [%d] from the RSS feed at %s", resp.StatusCode, url)
		return nil, err, resp.StatusCode
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		err := errors.Chain(fmt.Sprintf("failed to read the RSS feed at %s", url), err)
		return nil, err, http.StatusUnprocessableEntity
	}

	var rss RSS
	err = xml.Unmarshal(body, &rss)
	if err != nil {
		err := errors.Chain(fmt.Sprintf("failed to parse the RSS feed at %s", url), err)
		return nil, err, http.StatusInternalServerError
	}

	return &rss, nil, 0
}
