package server

import (
	"bytes"
	"encoding/json"
	"net"
	"net/http"
	"net/url"
	"strings"

	"morbo/context"
	"morbo/errors"
)

type feedHandler struct{}

func (conn *Connection) validateURL(rawURL string) error {
	const maximumURLLength = 2048

	if len(rawURL) > maximumURLLength {
		conn.Error("the URL is too long", http.StatusBadRequest)
		return errors.Err
	}

	url, err := url.Parse(rawURL)
	if err != nil {
		conn.Error("failed to parse the url", http.StatusBadRequest)
		return errors.Err
	}

	if !url.IsAbs() {
		conn.Error("only absolute URLs are supported", http.StatusBadRequest)
		return errors.Err
	}

	if url.Scheme != "https" && url.Scheme != "http" {
		conn.Error("this scheme is unsupported", http.StatusBadRequest)
		return errors.Err
	}

	host := url.Hostname()
	if host == "" {
		conn.Error("hostname cannot be empty", http.StatusBadRequest)
		return errors.Err
	}

	if strings.Count(host, ".") == 0 {
		conn.Error("internal network hostnames aren't allowed", http.StatusBadRequest)
		return errors.Err
	}

	port := url.Port()
	if port != "" && port != "80" && port != "443" {
		conn.Error("only standard HTTP(S) ports are allowed", http.StatusBadRequest)
		return errors.Err
	}

	ips, err := net.LookupIP(host)
	for _, ip := range ips {
		if ip.IsLoopback() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || ip.IsPrivate() {
			conn.Error("resolved IP is not allowed", http.StatusBadRequest)
			return errors.Err
		}
	}

	return nil
}

func (handler *feedHandler) handlePost(ctx context.Context, conn *Connection) error {
	sessionToken, err := conn.GetSessionToken()
	if err != nil {
		conn.log.Error.Println("failed to get the session token")
		return errors.Err
	}

	if _, err := conn.AuthenticateViaSessionToken(ctx, sessionToken); err != nil {
		conn.log.Error.Println("failed to authenticate via the session token")
		return errors.Err
	}

	type RequestBody struct {
		URL string `json:"url"`
	}

	var requestBody RequestBody
	if err := json.NewDecoder(conn.request.Body).Decode(&requestBody); err != nil {
		conn.log.Error.Println(err)
		conn.Error("failed to decode the request body", http.StatusBadRequest)
		return errors.Err
	}

	err = conn.validateURL(requestBody.URL)
	if err != nil {
		conn.log.Error.Println("failed to validate the URL")
		return errors.Err
	}

	rss, err := conn.parseRSS(ctx, requestBody.URL)
	if err != nil {
		conn.log.Error.Println("failed to parse the RSS feed")
		return errors.Err
	}

	type ResponseBody struct {
		Title string `json:"title"`
	}

	responseBody := ResponseBody{rss.Channel.Title}

	var responseBodyBuffer bytes.Buffer
	if err := json.NewEncoder(&responseBodyBuffer).Encode(&responseBody); err != nil {
		conn.DistinctError(
			"failed to encode the response",
			"internal server error",
			http.StatusInternalServerError,
		)
		return errors.Err
	}

	conn.writer.Header().Set("Content-Type", "application/json")

	if _, err = responseBodyBuffer.WriteTo(conn.writer); err != nil {
		conn.log.Error.Println("failed to write to the body")
		return errors.Err
	}

	return nil
}

func (handler *feedHandler) handleOptions(conn *Connection) {
	conn.writer.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
	conn.writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	conn.writer.WriteHeader(http.StatusOK)
}

func (handler *feedHandler) Handle(conn *Connection) {
	ctx := conn.request.Context()

	conn.log.Info.Printf(
		"%s %s %s\n",
		conn.request.RemoteAddr,
		conn.request.Method,
		conn.request.URL.Path,
	)

	switch conn.request.Method {
	case http.MethodPost:
		if err := handler.handlePost(ctx, conn); err != nil {
			conn.log.Error.Println("failed to handle the POST request to \"/feed/\"")
		}
	case http.MethodOptions:
		handler.handleOptions(conn)
	default:
		conn.writer.WriteHeader(http.StatusMethodNotAllowed)
	}
}
