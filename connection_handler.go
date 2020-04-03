package main

import (
	"context"
	"encoding/binary"
	"fmt"
	"github.com/flier/gohs/hyperscan"
	"github.com/google/gopacket"
	"github.com/google/gopacket/tcpassembly"
	"log"
	"sync"
	"time"
)

type BiDirectionalStreamFactory struct {
	storage      Storage
	serverIp     gopacket.Endpoint
	connections  map[StreamKey]ConnectionHandler
	mConnections sync.Mutex
	patterns     hyperscan.StreamDatabase
	mPatterns    sync.Mutex
	scratches    []*hyperscan.Scratch
}

type StreamKey [4]gopacket.Endpoint

type ConnectionHandler interface {
	Complete(handler *StreamHandler)
	Storage() Storage
	Context() context.Context
	Patterns() hyperscan.StreamDatabase
}

type connectionHandlerImpl struct {
	storage        Storage
	net, transport gopacket.Flow
	initiator      StreamKey
	connectionKey  string
	mComplete      sync.Mutex
	otherStream    *StreamHandler
	context        context.Context
	patterns       hyperscan.StreamDatabase
}

func (factory *BiDirectionalStreamFactory) New(net, transport gopacket.Flow) tcpassembly.Stream {
	key := StreamKey{net.Src(), net.Dst(), transport.Src(), transport.Dst()}
	invertedKey := StreamKey{net.Dst(), net.Src(), transport.Dst(), transport.Src()}

	factory.mConnections.Lock()
	connection, isPresent := factory.connections[invertedKey]
	if isPresent {
		delete(factory.connections, invertedKey)
	} else {
		var initiator StreamKey
		if net.Src() == factory.serverIp {
			initiator = invertedKey
		} else {
			initiator = key
		}
		connection = &connectionHandlerImpl{
			storage:   factory.storage,
			net:       net,
			transport: transport,
			initiator: initiator,
			mComplete: sync.Mutex{},
			context:   context.Background(),
			patterns : factory.patterns,
		}
		factory.connections[key] = connection
	}
	factory.mConnections.Unlock()

	streamHandler := NewStreamHandler(connection, key, factory.takeScratch())

	return &streamHandler
}

func (factory *BiDirectionalStreamFactory) UpdatePatternsDatabase(database hyperscan.StreamDatabase) {
	factory.mPatterns.Lock()
	factory.patterns = database

	for _, s := range factory.scratches {
		err := s.Realloc(database)
		if err != nil {
			fmt.Println("failed to realloc an existing scratch")
		}
	}

	factory.mPatterns.Unlock()
}

func (ch *connectionHandlerImpl) Complete(handler *StreamHandler) {
	ch.mComplete.Lock()
	if ch.otherStream == nil {
		ch.otherStream = handler
		ch.mComplete.Unlock()
		return
	}
	ch.mComplete.Unlock()

	var startedAt, closedAt time.Time
	if handler.firstPacketSeen.Before(ch.otherStream.firstPacketSeen) {
		startedAt = handler.firstPacketSeen
	} else {
		startedAt = ch.otherStream.firstPacketSeen
	}

	if handler.lastPacketSeen.After(ch.otherStream.lastPacketSeen) {
		closedAt = handler.lastPacketSeen
	} else {
		closedAt = ch.otherStream.lastPacketSeen
	}

	var client, server *StreamHandler
	if handler.streamKey == ch.initiator {
		client = handler
		server = ch.otherStream
	} else {
		client = ch.otherStream
		server = handler
	}

	ch.generateConnectionKey(startedAt)

	_, err := ch.storage.InsertOne(ch.context, "connections", OrderedDocument{
		{"_id", ch.connectionKey},
		{"ip_src", ch.initiator[0].String()},
		{"ip_dst", ch.initiator[1].String()},
		{"port_src", ch.initiator[2].String()},
		{"port_dst", ch.initiator[3].String()},
		{"started_at", startedAt},
		{"closed_at", closedAt},
		{"client_bytes", client.streamLength},
		{"server_bytes", server.streamLength},
		{"client_documents", len(client.documentsKeys)},
		{"server_documents", len(server.documentsKeys)},
		{"processed_at", time.Now()},
	})
	if err != nil {
		log.Println("error inserting document on collection connections with _id = ", ch.connectionKey)
	}

	streamsIds := append(client.documentsKeys, server.documentsKeys...)
	n, err := ch.storage.UpdateOne(ch.context, "connection_streams",
		UnorderedDocument{"_id": UnorderedDocument{"$in": streamsIds}},
		UnorderedDocument{"connection_id": ch.connectionKey},
		false)
	if err != nil {
		log.Println("failed to update connection streams", err)
	}
	if n != len(streamsIds) {
		log.Println("failed to update all connections streams")
	}
}

func (ch *connectionHandlerImpl) Storage() Storage {
	return ch.storage
}

func (ch *connectionHandlerImpl) Context() context.Context {
	return ch.context
}

func (ch *connectionHandlerImpl) Patterns() hyperscan.StreamDatabase {
	return ch.patterns
}

func (ch *connectionHandlerImpl) generateConnectionKey(firstPacketSeen time.Time) {
	hash := make([]byte, 16)
	binary.BigEndian.PutUint64(hash, uint64(firstPacketSeen.UnixNano()))
	binary.BigEndian.PutUint64(hash[8:], ch.net.FastHash()^ch.transport.FastHash())

	ch.connectionKey = fmt.Sprintf("%x", hash)
}

func (factory *BiDirectionalStreamFactory) takeScratch() *hyperscan.Scratch {
	factory.mPatterns.Lock()
	defer factory.mPatterns.Unlock()

	if len(factory.scratches) == 0 {
		scratch, err := hyperscan.NewScratch(factory.patterns)
		if err != nil {
			fmt.Println("failed to alloc a new scratch")
		}

		return scratch
	}

	index := len(factory.scratches) - 1
	scratch := factory.scratches[index]
	factory.scratches = factory.scratches[:index]

	return scratch
}

func (factory *BiDirectionalStreamFactory) releaseScratch(scratch *hyperscan.Scratch) {
	factory.mPatterns.Lock()
	factory.scratches = append(factory.scratches, scratch)
	factory.mPatterns.Unlock()
}
