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

func (conn *Connection) parseRSS(url string) (*RSS, error) {
	resp, err := http.Get(url)
	if err != nil {
		conn.Error("failed to request the RSS feed", http.StatusBadRequest)
		return nil, errors.Error
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		message := fmt.Sprintf("got a non-ok response code [%d]", resp.StatusCode)
		conn.Error(message, resp.StatusCode)
		return nil, errors.Error
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		conn.Error("failed to read the RSS feed", http.StatusUnprocessableEntity)
		return nil, errors.Error
	}

	var rss RSS
	err = xml.Unmarshal(body, &rss)
	if err != nil {
		conn.Error("failed to unmarshal the RSS feed", http.StatusUnprocessableEntity)
		return nil, errors.Error
	}

	return &rss, nil
}
