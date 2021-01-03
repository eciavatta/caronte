/*
 * This file is part of caronte (https://github.com/eciavatta/caronte).
 * Copyright (c) 2021 Emiliano Ciavatta.
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

package similarity

import (
	"bytes"
	"compress/zlib"
	"io"
)

func zlibDeflate(buffer []byte) []byte {
	bb := new(bytes.Buffer)
	zw := zlib.NewWriter(bb)
	if _, err := zw.Write(buffer); err != nil {
		panic(err)
	}
	if err := zw.Close(); err != nil {
		panic(err)
	}
	return bb.Bytes()
}

func zlibInflate(buffer []byte) []byte {
	var bb bytes.Buffer
	br := bytes.NewReader(buffer)
	if zr, err := zlib.NewReader(br); err == nil {
		if _, err = io.Copy(&bb, zr); err != nil {
			panic(err)
		}
		if err := zr.Close(); err != nil {
			panic(err)
		}
	} else {
		panic(err)
	}

	return bb.Bytes()
}

func min(a, b int) int {
	if a < b {
		return a
	} else {
		return b
	}
}

func max(a, b int) int {
	if a > b {
		return a
	} else {
		return b
	}
}
