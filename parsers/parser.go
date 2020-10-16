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

type Parser interface {
	TryParse(content []byte) Metadata

}

type Metadata interface {
}

type BasicMetadata struct {
	Type string `json:"type"`
}

var parsers = []Parser{	// order matter
	HTTPRequestParser{},
	HTTPResponseParser{},
}

func Parse(content []byte) Metadata {
	for _, parser := range parsers {
		if metadata := parser.TryParse(content); metadata != nil {
			return metadata
		}
	}

	return nil
}
