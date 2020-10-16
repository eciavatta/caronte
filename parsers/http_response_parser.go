/*
 * This file is part of caronte (https://github.com/eciavatta/caronte).
 * Copyright (c) 2020 Emiliano Ciavatta.
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, version 3.
 *
 * This program is distributed in the hope that it will be useful, but
 * WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU
 * General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program. If not, see <http://www.gnu.org/licenses/>.
 */

package parsers

import (
	"bufio"
	"bytes"
	"compress/gzip"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
)

type HTTPResponseMetadata struct {
	BasicMetadata
	Status           string            `json:"status"`
	StatusCode       int               `json:"status_code"`
	Protocol         string            `json:"protocol"`
	Headers          map[string]string `json:"headers"`
	ConnectionClosed bool              `json:"connection_closed"`
	Cookies          map[string]string `json:"cookies" binding:"omitempty"`
	Location         string            `json:"location" binding:"omitempty"`
	Compressed       bool              `json:"compressed"`
	Body             string            `json:"body" binding:"omitempty"`
	Trailer          map[string]string `json:"trailer" binding:"omitempty"`
}

type HTTPResponseParser struct {
}

func (p HTTPResponseParser) TryParse(content []byte) Metadata {
	reader := bufio.NewReader(bytes.NewReader(content))
	response, err := http.ReadResponse(reader, nil)
	if err != nil {
		return nil
	}
	var body string
	var compressed bool
	switch response.Header.Get("Content-Encoding") {
	case "gzip":
		if gzipReader, err := gzip.NewReader(response.Body); err == nil {
			if buffer, err := ioutil.ReadAll(gzipReader); err == nil {
				body = string(buffer)
				compressed = true
			} else {
				log.WithError(err).Error("failed to read gzipped body in http_response_parser")
				return nil
			}
			_ = gzipReader.Close()
		}
	default:
		if buffer, err := ioutil.ReadAll(response.Body); err == nil {
			body = string(buffer)
		} else {
			log.WithError(err).Error("failed to read body in http_response_parser")
			return nil
		}
	}
	_ = response.Body.Close()

	var location string
	if locationURL, err := response.Location(); err == nil {
		location = locationURL.String()
	}

	return HTTPResponseMetadata{
		BasicMetadata:    BasicMetadata{"http-response"},
		Status:           response.Status,
		StatusCode:       response.StatusCode,
		Protocol:         response.Proto,
		Headers:          JoinArrayMap(response.Header),
		ConnectionClosed: response.Close,
		Cookies:          CookiesMap(response.Cookies()),
		Location:         location,
		Compressed:       compressed,
		Body:             body,
		Trailer:          JoinArrayMap(response.Trailer),
	}
}
