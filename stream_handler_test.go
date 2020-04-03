package main

import (
	"context"
	"errors"
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
	streamHandler := createTestStreamHandler(t, testStorage{}, patterns)

	streamHandler.Reassembled([]tcpassembly.Reassembly{})
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
}

func TestReassemblingSingleDocumentStream(t *testing.T) {
	patterns, err := hyperscan.NewStreamDatabase(hyperscan.NewPattern("/impossible_to_match/", 0))
	require.Nil(t, err)
	storage := &testStorage{}
	streamHandler := createTestStreamHandler(t, storage, patterns)

	payloadLen := 256
	firstTime := time.Unix(1000000000, 0)
	middleTime := time.Unix(1000000010, 0)
	lastTime := time.Unix(1000000020, 0)
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
		assert.Equal(t, "bb41a60281cfae830000b6b3a7640000", od[0].Value)
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

	assert.Equal(t, data, streamHandler.buffer.Bytes(), "buffer should contains the same bytes of reassembles")
	assert.Equal(t, indexes, streamHandler.indexes, "indexes")
	assert.Equal(t, timestamps, streamHandler.timestamps, "timestamps")
	assert.Equal(t, lossBlocks, streamHandler.lossBlocks, "lossBlocks")
	assert.Equal(t, len(data), streamHandler.currentIndex)
	assert.Equal(t, firstTime, streamHandler.firstPacketSeen)
	assert.Equal(t, lastTime, streamHandler.lastPacketSeen)
	assert.Len(t, streamHandler.documentsKeys, 0)
	assert.Equal(t, len(data), streamHandler.streamLength)
	assert.Len(t, streamHandler.patternMatches, 0)

	completed := false
	streamHandler.connection.(*testConnectionHandler).onComplete = func(handler *StreamHandler) {
		completed = true
	}
	streamHandler.ReassemblyComplete()
	assert.Equal(t, true, inserted, "inserted")
	assert.Equal(t, true, completed, "completed")
}


func createTestStreamHandler(t *testing.T, storage Storage, patterns hyperscan.StreamDatabase) StreamHandler {
	testConnectionHandler := &testConnectionHandler{
		storage: storage,
		context: context.Background(),
		patterns: patterns,
	}

	scratch, err := hyperscan.NewScratch(patterns)
	require.Nil(t, err)

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
