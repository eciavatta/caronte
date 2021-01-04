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
	"testing"
)

func TestByteHistogram(t *testing.T) {
	data := []byte("Lorem ipsum dolor sit amet, consectetur adipiscing elit. Morbi ac sapien mi. " +
		"Duis nec pretium sapien, a accumsan nisi. Maecenas ac ligula augue. Pellentesque turpis orci, " +
		"faucibus sed rutrum in, elementum vitae diam. Vivamus laoreet dolor augue, " +
		"non convallis magna placerat bibendum. Nullam consequat erat nec nisl semper, vel eleifend magna eleifend. " +
		"Sed viverra augue quis lectus laoreet, quis consectetur eros hendrerit.")

	bh := ByteHistogramDigest(data)
	assert.NotNil(t, bh)
	emptyBc := ByteHistogramDigest([]byte{})
	assert.NotNil(t, emptyBc)
	assert.Zero(t, bh.Distance(bh))
	assert.Equal(t, len(data), bh.Distance(emptyBc))
	assert.Equal(t, len(data), emptyBc.Distance(bh))

	digest := bh.Digest()
	assert.Equal(t, bh.digest, ByteHistogramFromDigest(digest).digest)
	assert.Equal(t, bh.digest, ByteHistogramFromStringDigest(bh.StringDigest()).digest)
	assert.Equal(t, bh.Digest(), ByteHistogramFromDigest(digest).Digest())
	assert.Equal(t, bh.StringDigest(), ByteHistogramFromDigest(digest).StringDigest())
}
