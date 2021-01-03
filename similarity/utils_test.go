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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"math/rand"
	"testing"
)

func TestSimilarityUtils(t *testing.T) {
	// zlibDeflate, zlibInflate
	n := 4096
	buf := make([]byte, n)
	num, err := rand.Read(buf)
	require.Equal(t, n, num)
	require.NoError(t, err)

	compressed := zlibDeflate(buf)
	uncompressed := zlibInflate(compressed)

	assert.Equal(t, buf, uncompressed)

	// min
	assert.Equal(t, 0, min(0, 2))
	assert.Equal(t, -1, min(2, -1))

	// max
	assert.Equal(t, 2, max(0, 2))
	assert.Equal(t, 2, max(2, -1))
}
