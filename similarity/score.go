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

import "github.com/glaslos/tlsh"

const TlshMinSizeThreshold = 128
const ByteHistogramMaxSizeThreshold = 4096
const TlshSimilarityThreshold = 50
const ByteHistogramScoreThreshold = 0.9

type ComparableStream struct {
	Size          int
	Tlsh          *tlsh.Tlsh
	ByteHistogram *ByteHistogram
}

func (cs ComparableStream) IsSimilarTo(other ComparableStream) bool {
	minSize := min(cs.Size, other.Size)
	maxSize := max(cs.Size, other.Size)

	if cs.Tlsh != nil && other.Tlsh != nil && minSize > TlshMinSizeThreshold {
		if cs.Tlsh.Diff(other.Tlsh) < TlshSimilarityThreshold {
			return true
		}
	}

	if maxSize < ByteHistogramMaxSizeThreshold {
		score := float32(cs.ByteHistogram.Distance(other.ByteHistogram)) / float32(maxSize)
		if score < ByteHistogramScoreThreshold {
			return true
		}
	}

	return false
}
