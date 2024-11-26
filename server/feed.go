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
	"encoding/json"
	"net/http"

	"morbo/db"
	"morbo/errors"
	"morbo/log"
)

type feedHandler struct {
	db *db.DB
}

func (handler *feedHandler) handlePost(writer http.ResponseWriter, request *http.Request) error {
	type RequestBody struct {
		URL string `json:"url"`
	}

	var requestBody RequestBody
	if err := json.NewDecoder(request.Body).Decode(&requestBody); err != nil {
		log.Error.Println(err)
		Error(writer, "failed to decode the request body", http.StatusBadRequest)
		return errors.Error
	}

	rss, statusCode, err := parseRSS(requestBody.URL)
	if err != nil {
		Error(writer, "failed to parse the RSS feed", statusCode)
		return errors.Error
	}

	type ResponseBody struct {
		Title string `json:"title"`
	}
	writer.WriteHeader(http.StatusOK)
	json.NewEncoder(writer).Encode(ResponseBody{rss.Channel.Title})

	return nil
}

func (handler *feedHandler) handleOptions(writer http.ResponseWriter, _ *http.Request) {
	writer.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	writer.WriteHeader(http.StatusOK)
}

func (handler *feedHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	if origin := request.Header.Get("Origin"); origin != "" {
		writer.Header().Set("Access-Control-Allow-Origin", origin)
	}
	writer.Header().Set("Vary", "Origin")

	log.Info.Printf("%s %s %s\n", request.RemoteAddr, request.Method, request.URL.Path)
	switch request.Method {
	case http.MethodPost:
		if err := handler.handlePost(writer, request); err != nil {
			log.Error.Println("failed to handle the POST request to \"/feed/\"")
		}
	case http.MethodOptions:
		handler.handleOptions(writer, request)
	default:
		writer.WriteHeader(http.StatusMethodNotAllowed)
	}
}
