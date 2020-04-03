package main

import (
	"context"
	"errors"
	"fmt"
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

const testSrcIp = "10.10.10.100"
const testDstIp = "10.10.10.1"
const srcPort = 44444
const dstPort = 8080


func TestReassemblingEmptyStream(t *testing.T) {
	patterns, err := hyperscan.NewStreamDatabase(hyperscan.NewPattern("/nope/", 0))
	require.Nil(t, err)
	scratch, err := hyperscan.NewScratch(patterns)
	require.Nil(t, err)
	streamHandler := createTestStreamHandler(testStorage{}, patterns, scratch)

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
	assert.Len(t, streamHandler.documentsKeys, 0)
	assert.Zero(t, streamHandler.streamLength)
	assert.Len(t, streamHandler.patternMatches, 0)

	expected := 0
	streamHandler.connection.(*testConnectionHandler).onComplete = func(handler *StreamHandler) {
		expected = 42
	}
	streamHandler.ReassemblyComplete()
	assert.Equal(t, 42, expected)

	err = scratch.Free()
	require.Nil(t, err, "free scratch")
	err = patterns.Close()
	require.Nil(t, err, "close stream database")
}


func TestReassemblingSingleDocument(t *testing.T) {
	patterns, err := hyperscan.NewStreamDatabase(hyperscan.NewPattern("/impossible_to_match/", 0))
	require.Nil(t, err)
	scratch, err := hyperscan.NewScratch(patterns)
	require.Nil(t, err)
	storage := &testStorage{}
	streamHandler := createTestStreamHandler(storage, patterns, scratch)

	payloadLen := 256
	firstTime := time.Unix(0, 0)
	middleTime := time.Unix(10, 0)
	lastTime := time.Unix(20, 0)
	data := make([]byte, MaxDocumentSize)
	rand.Read(data)
	reassembles := make([]tcpassembly.Reassembly, MaxDocumentSize / payloadLen)
	indexes := make([]int, MaxDocumentSize / payloadLen)
	timestamps := make([]time.Time, MaxDocumentSize / payloadLen)
	lossBlocks := make([]bool, MaxDocumentSize / payloadLen)
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
			Bytes: data[i*payloadLen:(i+1)*payloadLen],
			Skip:  0,
			Start: i == 0,
			End:   i == len(reassembles)-1,
			Seen:  seen,
		}
		indexes[i] = i*payloadLen
		timestamps[i] = seen
	}

	inserted := false
	storage.insertFunc = func(ctx context.Context, collectionName string, document interface{}) (i interface{}, err error) {
		od := document.(OrderedDocument)
		assert.Equal(t, "connection_streams", collectionName)
		assert.Equal(t, "bb41a60281cfae830000000000000000", od[0].Value)
		assert.Equal(t, nil, od[1].Value)
		assert.Equal(t, 0, od[2].Value)
		assert.Equal(t, data, od[3].Value)
		assert.Equal(t, indexes, od[4].Value)
		assert.Equal(t, timestamps, od[5].Value)
		assert.Equal(t, lossBlocks, od[6].Value)
		assert.Len(t, od[7].Value, 0)
		inserted = true
		return nil, nil
	}

	streamHandler.Reassembled(reassembles)
	if !assert.Equal(t, false, inserted) {
		inserted = false
	}

	completed := false
	streamHandler.connection.(*testConnectionHandler).onComplete = func(handler *StreamHandler) {
		completed = true
	}
	streamHandler.ReassemblyComplete()

	assert.Equal(t, len(data), streamHandler.currentIndex)
	assert.Equal(t, firstTime, streamHandler.firstPacketSeen)
	assert.Equal(t, lastTime, streamHandler.lastPacketSeen)
	assert.Len(t, streamHandler.documentsKeys, 1)
	assert.Equal(t, len(data), streamHandler.streamLength)
	assert.Len(t, streamHandler.patternMatches, 0)

	assert.Equal(t, true, inserted, "inserted")
	assert.Equal(t, true, completed, "completed")

	err = scratch.Free()
	require.Nil(t, err, "free scratch")
	err = patterns.Close()
	require.Nil(t, err, "close stream database")
}


func TestReassemblingMultipleDocuments(t *testing.T) {
	patterns, err := hyperscan.NewStreamDatabase(hyperscan.NewPattern("/impossible_to_match/", 0))
	require.Nil(t, err)
	scratch, err := hyperscan.NewScratch(patterns)
	require.Nil(t, err)
	storage := &testStorage{}
	streamHandler := createTestStreamHandler(storage, patterns, scratch)

	payloadLen := 256
	firstTime := time.Unix(0, 0)
	middleTime := time.Unix(10, 0)
	lastTime := time.Unix(20, 0)
	dataSize := MaxDocumentSize*2
	data := make([]byte, dataSize)
	rand.Read(data)
	reassembles := make([]tcpassembly.Reassembly, dataSize / payloadLen)
	indexes := make([]int, dataSize / payloadLen)
	timestamps := make([]time.Time, dataSize / payloadLen)
	lossBlocks := make([]bool, dataSize / payloadLen)
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
			Bytes: data[i*payloadLen:(i+1)*payloadLen],
			Skip:  0,
			Start: i == 0,
			End:   i == len(reassembles)-1,
			Seen:  seen,
		}
		indexes[i] = i*payloadLen % MaxDocumentSize
		timestamps[i] = seen
	}

	inserted := 0
	storage.insertFunc = func(ctx context.Context, collectionName string, document interface{}) (i interface{}, err error) {
		od := document.(OrderedDocument)
		blockLen := MaxDocumentSize / payloadLen
		assert.Equal(t, "connection_streams", collectionName)
		assert.Equal(t, fmt.Sprintf("bb41a60281cfae83000%v000000000000", inserted), od[0].Value)
		assert.Equal(t, nil, od[1].Value)
		assert.Equal(t, inserted, od[2].Value)
		assert.Equal(t, data[MaxDocumentSize*inserted:MaxDocumentSize*(inserted+1)], od[3].Value)
		assert.Equal(t, indexes[blockLen*inserted:blockLen*(inserted+1)], od[4].Value)
		assert.Equal(t, timestamps[blockLen*inserted:blockLen*(inserted+1)], od[5].Value)
		assert.Equal(t, lossBlocks[blockLen*inserted:blockLen*(inserted+1)], od[6].Value)
		assert.Len(t, od[7].Value, 0)
		inserted += 1

		return nil, nil
	}

	streamHandler.Reassembled(reassembles)
	if !assert.Equal(t, 1, inserted) {
		inserted = 1
	}

	completed := false
	streamHandler.connection.(*testConnectionHandler).onComplete = func(handler *StreamHandler) {
		completed = true
	}
	streamHandler.ReassemblyComplete()

	assert.Equal(t, MaxDocumentSize, streamHandler.currentIndex)
	assert.Equal(t, firstTime, streamHandler.firstPacketSeen)
	assert.Equal(t, lastTime, streamHandler.lastPacketSeen)
	assert.Len(t, streamHandler.documentsKeys, 2)
	assert.Equal(t, len(data), streamHandler.streamLength)
	assert.Len(t, streamHandler.patternMatches, 0)

	assert.Equal(t, 2, inserted, "inserted")
	assert.Equal(t, true, completed, "completed")

	err = scratch.Free()
	require.Nil(t, err, "free scratch")
	err = patterns.Close()
	require.Nil(t, err, "close stream database")
}

func TestReassemblingPatternMatching(t *testing.T) {
	a, err := hyperscan.ParsePattern("/a{8}/i")
	require.Nil(t, err)
	a.Id = 0
	a.Flags |= hyperscan.SomLeftMost
	b, err := hyperscan.ParsePattern("/b[c]+b/i")
	require.Nil(t, err)
	b.Id = 1
	b.Flags |= hyperscan.SomLeftMost
	d, err := hyperscan.ParsePattern("/[d]+e[d]+/i")
	require.Nil(t, err)
	d.Id = 2
	d.Flags |= hyperscan.SomLeftMost

	payload := "aaaaaaaa0aaaaaaaaaa0bbbcccbbb0dddeddddedddd"
	expected := map[uint][]PatternSlice{
		0: {{0, 8}, {9, 17}, {10, 18}, {11, 19}},
		1: {{22, 27}},
		2: {{30, 38}, {34, 43}},
	}

	patterns, err := hyperscan.NewStreamDatabase(a, b, d)
	require.Nil(t, err)
	scratch, err := hyperscan.NewScratch(patterns)
	require.Nil(t, err)
	storage := &testStorage{}
	streamHandler := createTestStreamHandler(storage, patterns, scratch)

	seen := time.Unix(0, 0)
	inserted := false
	storage.insertFunc = func(ctx context.Context, collectionName string, document interface{}) (i interface{}, err error) {
		od := document.(OrderedDocument)
		assert.Equal(t, "connection_streams", collectionName)
		assert.Equal(t, "bb41a60281cfae830000000000000000", od[0].Value)
		assert.Equal(t, nil, od[1].Value)
		assert.Equal(t, 0, od[2].Value)
		assert.Equal(t, []byte(payload), od[3].Value)
		assert.Equal(t, []int{0}, od[4].Value)
		assert.Equal(t, []time.Time{seen}, od[5].Value)
		assert.Equal(t, []bool{false}, od[6].Value)
		assert.Equal(t, expected, od[7].Value)
		inserted = true

		return nil, nil
	}

	streamHandler.Reassembled([]tcpassembly.Reassembly{{
		Bytes: []byte(payload),
		Skip:  0,
		Start: true,
		End:   true,
		Seen:  seen,
	}})
	assert.Equal(t, false, inserted)

	completed := false
	streamHandler.connection.(*testConnectionHandler).onComplete = func(handler *StreamHandler) {
		completed = true
	}
	streamHandler.ReassemblyComplete()

	assert.Equal(t, len(payload), streamHandler.currentIndex)
	assert.Equal(t, seen, streamHandler.firstPacketSeen)
	assert.Equal(t, seen, streamHandler.lastPacketSeen)
	assert.Len(t, streamHandler.documentsKeys, 1)
	assert.Equal(t, len(payload), streamHandler.streamLength)

	assert.Equal(t, true, inserted, "inserted")
	assert.Equal(t, true, completed, "completed")

	err = scratch.Free()
	require.Nil(t, err, "free scratch")
	err = patterns.Close()
	require.Nil(t, err, "close stream database")
}


func createTestStreamHandler(storage Storage, patterns hyperscan.StreamDatabase, scratch *hyperscan.Scratch) StreamHandler {
	testConnectionHandler := &testConnectionHandler{
		storage: storage,
		context: context.Background(),
		patterns: patterns,
	}

	srcIp := layers.NewIPEndpoint(net.ParseIP(testSrcIp))
	dstIp := layers.NewIPEndpoint(net.ParseIP(testDstIp))
	srcPort := layers.NewTCPPortEndpoint(srcPort)
	dstPort := layers.NewTCPPortEndpoint(dstPort)

	return NewStreamHandler(testConnectionHandler, StreamKey{srcIp, dstIp, srcPort, dstPort}, scratch)
}

type testConnectionHandler struct {
	storage Storage
	context context.Context
	patterns hyperscan.StreamDatabase
	onComplete func(*StreamHandler)
}

func (tch *testConnectionHandler) Storage() Storage {
	return tch.storage
}

func (tch *testConnectionHandler) Context() context.Context {
	return tch.context
}

func (tch *testConnectionHandler) Patterns() hyperscan.StreamDatabase {
	return tch.patterns
}

func (tch *testConnectionHandler) Complete(handler *StreamHandler) {
	tch.onComplete(handler)
}

type testStorage struct {
	insertFunc func(ctx context.Context, collectionName string, document interface{}) (interface{}, error)
	updateOne func(ctx context.Context, collectionName string, filter interface{}, update interface {}, upsert bool) (interface{}, error)
	findOne func(ctx context.Context, collectionName string, filter interface{}) (UnorderedDocument, error)
}

func (ts testStorage) InsertOne(ctx context.Context, collectionName string, document interface{}) (interface{}, error) {
	if ts.insertFunc != nil {
		return ts.insertFunc(ctx, collectionName, document)
	}
	return nil, errors.New("not implemented")
}

func (ts testStorage) UpdateOne(ctx context.Context, collectionName string, filter interface{}, update interface {},
	upsert bool) (interface{}, error) {
	if ts.updateOne != nil {
		return ts.updateOne(ctx, collectionName, filter, update, upsert)
	}
	return nil, errors.New("not implemented")
}

func (ts testStorage) FindOne(ctx context.Context, collectionName string, filter interface{}) (UnorderedDocument, error) {
	if ts.insertFunc != nil {
		return ts.findOne(ctx, collectionName, filter)
	}
	return nil, errors.New("not implemented")
}
