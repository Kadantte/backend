// Copyright (C) 2024 Pavel Sobolev
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published
// by the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package server

import (
	"encoding/xml"
	"io"
	"net/http"

	"morbo/errors"
	"morbo/log"
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

func parseRSS(url string) (*RSS, statusCode, error) {
	resp, err := http.Get(url)
	if err != nil {
		log.Error.Println(err)
		log.Error.Println("failed to request the RSS feed at", url)
		return nil, http.StatusBadRequest, errors.Error
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Error.Printf("got a non-ok response code [%d] from the RSS feed at %s", resp.StatusCode, url)
		return nil, resp.StatusCode, errors.Error
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error.Println(err)
		log.Error.Printf("failed to read the RSS feed at %s", url)
		return nil, http.StatusUnprocessableEntity, errors.Error
	}

	var rss RSS
	err = xml.Unmarshal(body, &rss)
	if err != nil {
		log.Error.Println(err)
		log.Error.Printf("failed to parse the RSS feed at %s", url)
		return nil, http.StatusInternalServerError, errors.Error
	}

	return &rss, 0, nil
}
