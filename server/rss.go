package server

import (
	"encoding/xml"
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
	if !conn.ContextAlive() {
		return nil, errors.Error
	}

	resp, err := http.Get(url)
	if err != nil {
		conn.Error("failed to request the resource", http.StatusBadRequest)
		return nil, errors.Error
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		switch resp.StatusCode {
		case http.StatusNotFound:
			conn.Error("couldn't find the resource", http.StatusNotFound)
		case http.StatusForbidden:
			conn.Error("the resource is forbidden", http.StatusForbidden)
		default:
			conn.Error("the resource is not available", resp.StatusCode)
		}
		return nil, errors.Error
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		conn.Error("failed to read the resource", http.StatusUnprocessableEntity)
		return nil, errors.Error
	}

	var rss RSS
	err = xml.Unmarshal(body, &rss)
	if err != nil {
		conn.Error("failed to parse the resource as an RSS feed", http.StatusUnprocessableEntity)
		return nil, errors.Error
	}

	return &rss, nil
}
