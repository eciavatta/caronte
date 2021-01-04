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
	"encoding/hex"
	"github.com/glaslos/tlsh"
)

func swap(data byte) byte {
	result := ((data & 0xF0) >> 4) & 0x0F
	result |= ((data & 0x0F) << 4) & 0xF0
	return result
}

func ParseTlshDigest(digest string) *tlsh.Tlsh {
	buf, err := hex.DecodeString(digest)
	if err != nil || len(buf) != 35 {
		return nil
	}

	checksum := swap(buf[0])
	lValue := swap(buf[1])
	qRatio := buf[2]
	q1Ratio := qRatio >> 4
	q2Ratio := qRatio & 0xF
	code := [32]byte{}
	for i := 3; i < len(buf); i++ {
		code[i-3] = buf[i]
	}

	return tlsh.New(checksum, lValue, q1Ratio, q2Ratio, qRatio, code)
}
