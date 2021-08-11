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
