package parsers

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"io/ioutil"
	"net/http"
)

type HttpResponseMetadata struct {
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

type HttpResponseParser struct {
}

func (p HttpResponseParser) TryParse(content []byte) Metadata {
	reader := bufio.NewReader(bytes.NewReader(content))
	response, err := http.ReadResponse(reader, nil)
	if err != nil {
		return nil
	}
	var body string
	var compressed bool
	if response.Body != nil {
		switch response.Header.Get("Content-Encoding") {
		case "gzip":
			if gzipReader, err := gzip.NewReader(response.Body); err == nil {
				if buffer, err := ioutil.ReadAll(gzipReader); err == nil {
					body = string(buffer)
					compressed = true
				}
				_ = gzipReader.Close()
			}
		default:
			if buffer, err := ioutil.ReadAll(response.Body); err == nil {
				body = string(buffer)
			}
		}
		_ = response.Body.Close()
	}

	var location string
	if locationUrl, err := response.Location(); err == nil {
		location = locationUrl.String()
	}

	return HttpResponseMetadata{
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
