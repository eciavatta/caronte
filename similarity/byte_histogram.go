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
	"encoding/base64"
)

type ByteHistogram struct {
	digest []byte
}

func ByteHistogramFromStringDigest(digest string) *ByteHistogram {
	if buf, err := base64.StdEncoding.DecodeString(digest); err == nil {
		return ByteHistogramFromDigest(buf)
	} else {
		return nil
	}
}

func ByteHistogramFromDigest(digest []byte) *ByteHistogram {
	return &ByteHistogram{digest: zlibInflate(digest)}
}

func ByteHistogramDigest(buffer []byte) *ByteHistogram {
	digest := make([]byte, 256)
	for _, b := range buffer {
		if digest[b] < 255 {
			digest[b] += 1
		}
	}

	return &ByteHistogram{digest}
}

func (bc *ByteHistogram) StringDigest() string {
	digest := bc.Digest()
	buf := make([]byte, base64.StdEncoding.EncodedLen(len(digest)))
	base64.StdEncoding.Encode(buf, digest)
	return string(buf)
}

func (bc *ByteHistogram) Digest() []byte {
	return zlibDeflate(bc.digest)
}

func (bc *ByteHistogram) Distance(other *ByteHistogram) int {
	sum := 0
	for i := range bc.digest {
		if bc.digest[i] > other.digest[i] {
			sum += int(bc.digest[i] - other.digest[i])
		} else {
			sum += int(other.digest[i] - bc.digest[i])
		}
	}
	return sum
}
