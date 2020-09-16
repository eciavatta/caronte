package parsers

import (
	"net/http"
	"strings"
)

func JoinArrayMap(obj map[string][]string) map[string]string {
	headers := make(map[string]string, len(obj))
	for key, value := range obj {
		headers[key] = strings.Join(value, ";")
	}

	return headers
}

func CookiesMap(cookiesArray []*http.Cookie) map[string]string {
	cookies := make(map[string]string, len(cookiesArray))
	for _, cookie := range cookiesArray {
		cookies[cookie.Name] = cookie.Value
	}

	return cookies
}
