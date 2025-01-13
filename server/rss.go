package server

import (
	"encoding/xml"
	"io"
	"net/http"

	"morbo/context"
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

func (conn *Connection) parseRSS(ctx context.Context, url string) (*RSS, error) {
	if !conn.ContextAlive(ctx) {
		return nil, errors.Err
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		conn.DistinctError(
			"failed to prepare a request to the resource",
			"internal server error",
			http.StatusInternalServerError,
		)
		return nil, errors.Err
	}

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		conn.Error("failed to request the resource", http.StatusBadRequest)
		return nil, errors.Err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		switch response.StatusCode {
		case http.StatusNotFound:
			conn.Error("couldn't find the resource", http.StatusNotFound)
		case http.StatusForbidden:
			conn.Error("the resource is forbidden", http.StatusForbidden)
		default:
			conn.Error("the resource is not available", response.StatusCode)
		}
		return nil, errors.Err
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		conn.Error("failed to read the resource", http.StatusUnprocessableEntity)
		return nil, errors.Err
	}

	var rss RSS
	if err := xml.Unmarshal(body, &rss); err != nil {
		conn.Error("failed to parse the resource as an RSS feed", http.StatusUnprocessableEntity)
		return nil, errors.Err
	}

	return &rss, nil
}
