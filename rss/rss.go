package rss

import (
	"context"
	"encoding/xml"
	"io"
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

type ErrorTag interface {
	~uint64
}

type Error[T ErrorTag] struct {
	Tag  T
	Data any
}

type ParseErrorTag uint64

const (
	ErrRequestNew ParseErrorTag = iota
	ErrRequestSend
	ErrRequestNotOk
	ErrResponseRead
	ErrResponseParse
)

type ParseError Error[ParseErrorTag]

func ParseRSS(ctx context.Context, url string) (*RSS, *ParseError) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, &ParseError{ErrRequestNew, nil}
	}

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, &ParseError{ErrRequestSend, nil}
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, &ParseError{ErrRequestNotOk, &response.StatusCode}
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, &ParseError{ErrResponseRead, nil}
	}

	var rss RSS
	if err := xml.Unmarshal(body, &rss); err != nil {
		return nil, &ParseError{ErrResponseParse, nil}
	}

	return &rss, nil
}
