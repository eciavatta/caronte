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
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"moul.io/http2curl"
	"net/http"
	"strings"
)

type HTTPRequestMetadata struct {
	BasicMetadata
	Method        string                         `json:"method"`
	URL           string                         `json:"url"`
	Protocol      string                         `json:"protocol"`
	Host          string                         `json:"host"`
	Headers       map[string]string              `json:"headers"`
	Cookies       map[string]string              `json:"cookies" binding:"omitempty"`
	ContentLength int64                          `json:"content_length"`
	FormData      map[string]string              `json:"form_data" binding:"omitempty"`
	Body          string                         `json:"body" binding:"omitempty"`
	Trailer       map[string]string              `json:"trailer" binding:"omitempty"`
	Reproducers   HTTPRequestMetadataReproducers `json:"reproducers"`
}

type HTTPRequestMetadataReproducers struct {
	CurlCommand  string `json:"curl_command"`
	RequestsCode string `json:"requests_code"`
	FetchRequest string `json:"fetch_request"`
}

type HTTPRequestParser struct {
}

func (p HTTPRequestParser) TryParse(content []byte) Metadata {
	reader := bufio.NewReader(bytes.NewReader(content))
	request, err := http.ReadRequest(reader)
	if err != nil {
		return nil
	}
	var body string
	if buffer, err := ioutil.ReadAll(request.Body); err == nil {
		body = string(buffer)
	} else {
		log.WithError(err).Error("failed to read body in http_request_parser")
		return nil
	}
	_ = request.Body.Close()
	_ = request.ParseForm()

	return HTTPRequestMetadata{
		BasicMetadata: BasicMetadata{"http-request"},
		Method:        request.Method,
		URL:           request.URL.String(),
		Protocol:      request.Proto,
		Host:          request.Host,
		Headers:       JoinArrayMap(request.Header),
		Cookies:       CookiesMap(request.Cookies()),
		ContentLength: request.ContentLength,
		FormData:      JoinArrayMap(request.Form),
		Body:          body,
		Trailer:       JoinArrayMap(request.Trailer),
		Reproducers: HTTPRequestMetadataReproducers{
			CurlCommand:  curlCommand(content),
			RequestsCode: requestsCode(request, body),
			FetchRequest: fetchRequest(request, body),
		},
	}
}

func curlCommand(content []byte) string {
	// a new reader is required because all the body is read before and GetBody() doesn't works
	reader := bufio.NewReader(bytes.NewReader(content))
	request, _ := http.ReadRequest(reader)
	command, err := http2curl.GetCurlCommand(request)
	if err == nil {
		return command.String()
	}
	return err.Error()
}

func requestsCode(request *http.Request, body string) string {
	var b strings.Builder
	headers := toJSON(JoinArrayMap(request.Header))
	cookies := toJSON(CookiesMap(request.Cookies()))

	b.WriteString("import requests\n\nresponse = requests." + strings.ToLower(request.Method) + "(")
	b.WriteString("\"" + request.URL.String() + "\"")
	if body != "" {
		b.WriteString(", data = \"" + strings.Replace(body, "\"", "\\\"", -1) + "\"")
	}
	if headers != "" {
		b.WriteString(", headers = " + headers)
	}
	if cookies != "" {
		b.WriteString(", cookies = " + cookies)
	}
	b.WriteString(")\n")
	b.WriteString(`
# print(response.url)
# print(response.text)
# print(response.content)
# print(response.json())
# print(response.raw)
# print(response.status_code)
# print(response.cookies)
# print(response.history)
`)

	return b.String()
}

func fetchRequest(request *http.Request, body string) string {
	headers := JoinArrayMap(request.Header)
	data := make(map[string]interface{})
	data["headers"] = headers
	if referrer := request.Header.Get("referrer"); referrer != "" {
		data["Referrer"] = referrer
	}
	// TODO: referrerPolicy
	if body == "" {
		data["body"] = nil
	} else {
		data["body"] = body
	}
	data["method"] = request.Method
	// TODO: mode

	if jsonData := toJSON(data); jsonData != "" {
		return "fetch(\"" + request.URL.String() + "\", " + jsonData + ");"
	}
	return "invalid-request"
}

func toJSON(obj interface{}) string {
	if buffer, err := json.MarshalIndent(obj, "", "\t"); err == nil {
		return string(buffer)
	} else {
		return ""
	}
}
