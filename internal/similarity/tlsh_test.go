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
	"github.com/glaslos/tlsh"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func parseDigest(fileName, expectedHash string, t *testing.T) {
	expected, err := tlsh.HashFilename(fileName)
	require.NoError(t, err)

	actual := ParseTlshDigest(expectedHash)

	assert.Equal(t, expected.Binary(), actual.Binary())
	assert.Equal(t, expected.String(), actual.String())
}

func TestParseTlshDigest(t *testing.T) {
	parseDigest("../test_data/icmp.pcap",
		"B524E284C2E5C8EBDDD731FCE5E6D1DB238361953284C134FE364BD9896A17AC9A281C", t)
	parseDigest("../test_data/ping_pong_10000.pcap",
		"D745E1E0AB73862AF5D7AEB5D9D0F1E3179AB1D80F30F410EC095615521A0BD8BDEB21", t)
}
