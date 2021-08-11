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

package main

import (
	"context"
	"github.com/flier/gohs/hyperscan"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/tcpassembly"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"math/rand"
	"net"
	"testing"
	"time"
)

const testSrcIP = "10.10.10.100"
const testDstIP = "10.10.10.1"
const srcPort = 44444
const dstPort = 8080

func TestReassemblingEmptyStream(t *testing.T) {
	wrapper := NewTestStorageWrapper(t)
	wrapper.AddCollection(ConnectionStreams)
	patterns, err := hyperscan.NewStreamDatabase(hyperscan.NewPattern("/nope/", 0))
	require.NoError(t, err)
	scratch, err := hyperscan.NewScratch(patterns)
	require.NoError(t, err)
	streamHandler := createTestStreamHandler(wrapper, patterns, scratch)

	streamHandler.Reassembled([]tcpassembly.Reassembly{{
		Bytes: []byte{},
		Skip:  0,
		Start: true,
		End:   true,
	}})
	assert.Len(t, streamHandler.indexes, 0, "indexes")
	assert.Len(t, streamHandler.timestamps, 0, "timestamps")
	assert.Len(t, streamHandler.lossBlocks, 0)
	assert.Zero(t, streamHandler.currentIndex)
	assert.Zero(t, streamHandler.firstPacketSeen)
	assert.Zero(t, streamHandler.lastPacketSeen)
	assert.Len(t, streamHandler.documentsIDs, 0)
	assert.Zero(t, streamHandler.streamLength)
	assert.Len(t, streamHandler.patternMatches, 0)

	completed := false
	streamHandler.connection.(*testConnectionHandler).onComplete = func(handler *StreamHandler) {
		completed = true
	}
	streamHandler.ReassemblyComplete()
	assert.Equal(t, true, completed)

	err = scratch.Free()
	require.NoError(t, err, "free scratch")
	err = patterns.Close()
	require.NoError(t, err, "close stream database")
	wrapper.Destroy(t)
}

func TestReassemblingSingleDocument(t *testing.T) {
	wrapper := NewTestStorageWrapper(t)
	wrapper.AddCollection(ConnectionStreams)
	patterns, err := hyperscan.NewStreamDatabase(hyperscan.NewPattern("/impossible_to_match/", 0))
	require.NoError(t, err)
	scratch, err := hyperscan.NewScratch(patterns)
	require.NoError(t, err)
	streamHandler := createTestStreamHandler(wrapper, patterns, scratch)

	payloadLen := 256
	firstTime := time.Unix(0, 0)
	middleTime := time.Unix(10, 0)
	lastTime := time.Unix(20, 0)
	data := make([]byte, MaxDocumentSize)
	rand.Read(data)
	reassembles := make([]tcpassembly.Reassembly, MaxDocumentSize/payloadLen)
	indexes := make([]int, MaxDocumentSize/payloadLen)
	timestamps := make([]time.Time, MaxDocumentSize/payloadLen)
	lossBlocks := make([]bool, MaxDocumentSize/payloadLen)
	for i := 0; i < len(reassembles); i++ {
		var seen time.Time
		if i == 0 {
			seen = firstTime
		} else if i == len(reassembles)-1 {
			seen = lastTime
		} else {
			seen = middleTime
		}

		reassembles[i] = tcpassembly.Reassembly{
			Bytes: data[i*payloadLen : (i+1)*payloadLen],
			Skip:  0,
			Start: i == 0,
			End:   i == len(reassembles)-1,
			Seen:  seen,
		}
		indexes[i] = i * payloadLen
		timestamps[i] = seen
	}

	var results []ConnectionStream

	streamHandler.Reassembled(reassembles)
	err = wrapper.Storage.Find(ConnectionStreams).Context(wrapper.Context).All(&results)
	require.NoError(t, err)
	assert.Len(t, results, 0)

	completed := false
	streamHandler.connection.(*testConnectionHandler).onComplete = func(handler *StreamHandler) {
		completed = true
	}
	streamHandler.ReassemblyComplete()

	err = wrapper.Storage.Find(ConnectionStreams).Context(wrapper.Context).All(&results)
	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, firstTime.Unix(), results[0].ID.Timestamp().Unix())
	assert.Zero(t, results[0].ConnectionID)
	assert.Equal(t, 0, results[0].DocumentIndex)
	assert.Equal(t, data, results[0].Payload)
	assert.Equal(t, indexes, results[0].BlocksIndexes)
	assert.Len(t, results[0].BlocksTimestamps, len(timestamps)) // should be compared one by one
	assert.Equal(t, lossBlocks, results[0].BlocksLoss)
	assert.Len(t, results[0].PatternMatches, 0)

	assert.Equal(t, len(data), streamHandler.currentIndex)
	assert.Equal(t, firstTime, streamHandler.firstPacketSeen)
	assert.Equal(t, lastTime, streamHandler.lastPacketSeen)
	assert.Len(t, streamHandler.documentsIDs, 1)
	assert.Equal(t, len(data), streamHandler.streamLength)
	assert.Len(t, streamHandler.patternMatches, 0)

	assert.Equal(t, true, completed, "completed")

	err = scratch.Free()
	require.NoError(t, err, "free scratch")
	err = patterns.Close()
	require.NoError(t, err, "close stream database")
	wrapper.Destroy(t)
}

func TestReassemblingMultipleDocuments(t *testing.T) {
	wrapper := NewTestStorageWrapper(t)
	wrapper.AddCollection(ConnectionStreams)
	patterns, err := hyperscan.NewStreamDatabase(hyperscan.NewPattern("/impossible_to_match/", 0))
	require.NoError(t, err)
	scratch, err := hyperscan.NewScratch(patterns)
	require.NoError(t, err)
	streamHandler := createTestStreamHandler(wrapper, patterns, scratch)

	payloadLen := 256
	firstTime := time.Unix(0, 0)
	middleTime := time.Unix(10, 0)
	lastTime := time.Unix(20, 0)
	dataSize := MaxDocumentSize * 2
	data := make([]byte, dataSize)
	rand.Read(data)
	reassembles := make([]tcpassembly.Reassembly, dataSize/payloadLen)
	indexes := make([]int, dataSize/payloadLen)
	timestamps := make([]time.Time, dataSize/payloadLen)
	lossBlocks := make([]bool, dataSize/payloadLen)
	for i := 0; i < len(reassembles); i++ {
		var seen time.Time
		if i == 0 {
			seen = firstTime
		} else if i == len(reassembles)-1 {
			seen = lastTime
		} else {
			seen = middleTime
		}

		reassembles[i] = tcpassembly.Reassembly{
			Bytes: data[i*payloadLen : (i+1)*payloadLen],
			Skip:  0,
			Start: i == 0,
			End:   i == len(reassembles)-1,
			Seen:  seen,
		}
		indexes[i] = i * payloadLen % MaxDocumentSize
		timestamps[i] = seen
	}

	streamHandler.Reassembled(reassembles)

	var results []ConnectionStream
	err = wrapper.Storage.Find(ConnectionStreams).Context(wrapper.Context).All(&results)
	require.NoError(t, err)
	assert.Len(t, results, 1)

	completed := false
	streamHandler.connection.(*testConnectionHandler).onComplete = func(handler *StreamHandler) {
		completed = true
	}
	streamHandler.ReassemblyComplete()

	err = wrapper.Storage.Find(ConnectionStreams).Context(wrapper.Context).All(&results)
	require.NoError(t, err)
	assert.Len(t, results, 2)
	for i := 0; i < 2; i++ {
		blockLen := MaxDocumentSize / payloadLen
		assert.Equal(t, firstTime.Unix(), results[i].ID.Timestamp().Unix())
		assert.Zero(t, results[i].ConnectionID)
		assert.Equal(t, i, results[i].DocumentIndex)
		assert.Equal(t, data[MaxDocumentSize*i:MaxDocumentSize*(i+1)], results[i].Payload)
		assert.Equal(t, indexes[blockLen*i:blockLen*(i+1)], results[i].BlocksIndexes)
		assert.Len(t, results[i].BlocksTimestamps, len(timestamps[blockLen*i:blockLen*(i+1)])) // should be compared one by one
		assert.Equal(t, lossBlocks[blockLen*i:blockLen*(i+1)], results[i].BlocksLoss)
		assert.Len(t, results[i].PatternMatches, 0)
	}

	assert.Equal(t, MaxDocumentSize, streamHandler.currentIndex)
	assert.Equal(t, firstTime, streamHandler.firstPacketSeen)
	assert.Equal(t, lastTime, streamHandler.lastPacketSeen)
	assert.Len(t, streamHandler.documentsIDs, 2)
	assert.Equal(t, len(data), streamHandler.streamLength)
	assert.Len(t, streamHandler.patternMatches, 0)

	assert.Equal(t, true, completed, "completed")

	err = scratch.Free()
	require.NoError(t, err, "free scratch")
	err = patterns.Close()
	require.NoError(t, err, "close stream database")
	wrapper.Destroy(t)
}

func TestReassemblingPatternMatching(t *testing.T) {
	wrapper := NewTestStorageWrapper(t)
	wrapper.AddCollection(ConnectionStreams)
	a, err := hyperscan.ParsePattern("/a{8}/i")
	require.NoError(t, err)
	a.Id = 0
	a.Flags |= hyperscan.SomLeftMost
	b, err := hyperscan.ParsePattern("/b[c]+b/i")
	require.NoError(t, err)
	b.Id = 1
	b.Flags |= hyperscan.SomLeftMost
	d, err := hyperscan.ParsePattern("/[d]+e[d]+/i")
	require.NoError(t, err)
	d.Id = 2
	d.Flags |= hyperscan.SomLeftMost

	payload := "aaaaaaaa0aaaaaaaaaa0bbbcccbbb0dddeddddedddd"
	expected := map[uint][]PatternSlice{
		0: {{0, 8}, {9, 17}, {10, 18}, {11, 19}},
		1: {{22, 27}},
		2: {{30, 38}, {34, 43}},
	}

	patterns, err := hyperscan.NewStreamDatabase(a, b, d)
	require.NoError(t, err)
	scratch, err := hyperscan.NewScratch(patterns)
	require.NoError(t, err)
	streamHandler := createTestStreamHandler(wrapper, patterns, scratch)

	seen := time.Unix(0, 0)
	streamHandler.Reassembled([]tcpassembly.Reassembly{{
		Bytes: []byte(payload),
		Skip:  0,
		Start: true,
		End:   true,
		Seen:  seen,
	}})

	var results []ConnectionStream
	err = wrapper.Storage.Find(ConnectionStreams).Context(wrapper.Context).All(&results)
	require.NoError(t, err)
	assert.Len(t, results, 0)

	completed := false
	streamHandler.connection.(*testConnectionHandler).onComplete = func(handler *StreamHandler) {
		completed = true
	}
	streamHandler.ReassemblyComplete()

	err = wrapper.Storage.Find(ConnectionStreams).Context(wrapper.Context).All(&results)
	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, seen.Unix(), results[0].ID.Timestamp().Unix())
	assert.Zero(t, results[0].ConnectionID)
	assert.Equal(t, 0, results[0].DocumentIndex)
	assert.Equal(t, []byte(payload), results[0].Payload)
	assert.Equal(t, []int{0}, results[0].BlocksIndexes)
	assert.Len(t, results[0].BlocksTimestamps, 1) // should be compared one by one
	assert.Equal(t, []bool{false}, results[0].BlocksLoss)
	assert.Equal(t, expected, results[0].PatternMatches)

	assert.Equal(t, len(payload), streamHandler.currentIndex)
	assert.Equal(t, seen, streamHandler.firstPacketSeen)
	assert.Equal(t, seen, streamHandler.lastPacketSeen)
	assert.Len(t, streamHandler.documentsIDs, 1)
	assert.Equal(t, len(payload), streamHandler.streamLength)

	assert.Equal(t, true, completed, "completed")

	err = scratch.Free()
	require.NoError(t, err, "free scratch")
	err = patterns.Close()
	require.NoError(t, err, "close stream database")
	wrapper.Destroy(t)
}

func createTestStreamHandler(wrapper *TestStorageWrapper, patterns hyperscan.StreamDatabase, scratch *hyperscan.Scratch) StreamHandler {
	testConnectionHandler := &testConnectionHandler{
		wrapper:  wrapper,
		patterns: patterns,
	}

	srcIP := layers.NewIPEndpoint(net.ParseIP(testSrcIP))
	dstIP := layers.NewIPEndpoint(net.ParseIP(testDstIP))
	srcPort := layers.NewTCPPortEndpoint(srcPort)
	dstPort := layers.NewTCPPortEndpoint(dstPort)

	scanner := Scanner{scratch: scratch, version: ZeroRowID}
	return NewStreamHandler(testConnectionHandler, StreamFlow{srcIP, dstIP, srcPort, dstPort}, scanner, true) // TODO: test isClient
}

type testConnectionHandler struct {
	wrapper    *TestStorageWrapper
	patterns   hyperscan.StreamDatabase
	onComplete func(*StreamHandler)
}

func (tch *testConnectionHandler) Storage() Storage {
	return tch.wrapper.Storage
}

func (tch *testConnectionHandler) Context() context.Context {
	return tch.wrapper.Context
}

func (tch *testConnectionHandler) PatternsDatabase() hyperscan.StreamDatabase {
	return tch.patterns
}

func (tch *testConnectionHandler) PatternsDatabaseSize() int {
	return 8
}

func (tch *testConnectionHandler) Complete(handler *StreamHandler) {
	tch.onComplete(handler)
}
