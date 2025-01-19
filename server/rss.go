package server

import (
	"net/http"

	"morbo/context"
	"morbo/errors"
	"morbo/rss"
)

func (conn *Connection) parseRSS(ctx context.Context, url string) (*rss.RSS, error) {
	if !conn.ContextAlive(ctx) {
		return nil, errors.Err
	}

	feed, err := rss.ParseRSS(ctx, url)
	if err != nil {
		switch err.Tag {
		case rss.ErrRequestNew:
			conn.DistinctError(
				"failed to prepare a request to the resource",
				"internal server error",
				http.StatusInternalServerError,
			)
		case rss.ErrRequestSend:
			conn.Error("failed to request the resource", http.StatusBadRequest)
		case rss.ErrRequestNotOk:
			statusCode, ok := err.Data.(*int)
			if !ok {
				conn.Error("the resource is not available", http.StatusBadRequest)
			}
			switch *statusCode {
			case http.StatusUnauthorized:
				conn.Error("access to the resource is unauthorized", http.StatusForbidden)
			case http.StatusForbidden:
				conn.Error("access to the resource is forbidden", http.StatusForbidden)
			case http.StatusNotFound:
				conn.Error("couldn't find the resource", http.StatusNotFound)
			default:
				conn.Error("the resource is not available", *statusCode)
			}
		case rss.ErrResponseRead:
			conn.Error("failed to read the resource", http.StatusUnprocessableEntity)
		case rss.ErrResponseParse:
			conn.Error("failed to parse the resource as an RSS feed", http.StatusUnprocessableEntity)
		}
		return nil, errors.Err
	}

	return feed, nil
}
