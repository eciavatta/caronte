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
	"bytes"
	"context"
	"fmt"
	"github.com/eciavatta/caronte/parsers"
	log "github.com/sirupsen/logrus"
	"strings"
	"time"
)

const (
	initialMessagesSize     = 1024
	initialRegexSlicesCount = 8
	pwntoolsMaxServerBytes  = 20
)

type ConnectionStream struct {
	ID               RowID                   `bson:"_id"`
	ConnectionID     RowID                   `bson:"connection_id"`
	FromClient       bool                    `bson:"from_client"`
	DocumentIndex    int                     `bson:"document_index"`
	Payload          []byte                  `bson:"payload"`
	PayloadString    string                  `bson:"payload_string"`
	BlocksIndexes    []int                   `bson:"blocks_indexes"`
	BlocksTimestamps []time.Time             `bson:"blocks_timestamps"`
	BlocksLoss       []bool                  `bson:"blocks_loss"`
	PatternMatches   map[uint][]PatternSlice `bson:"pattern_matches"`
}

type PatternSlice [2]uint64

type Message struct {
	FromClient             bool             `json:"from_client"`
	Content                string           `json:"content"`
	Metadata               parsers.Metadata `json:"metadata"`
	IsMetadataContinuation bool             `json:"is_metadata_continuation"`
	Index                  int              `json:"index"`
	Timestamp              time.Time        `json:"timestamp"`
	IsRetransmitted        bool             `json:"is_retransmitted"`
	RegexMatches           []RegexSlice     `json:"regex_matches"`
}

type RegexSlice struct {
	From uint64 `json:"from"`
	To   uint64 `json:"to"`
}

type GetMessageFormat struct {
	Format string `form:"format"`
}

type DownloadMessageFormat struct {
	Format string `form:"format"`
	Type   string `form:"type"`
}

type ConnectionStreamsController struct {
	storage Storage
}

func NewConnectionStreamsController(storage Storage) ConnectionStreamsController {
	return ConnectionStreamsController{
		storage: storage,
	}
}

func (csc ConnectionStreamsController) GetConnectionMessages(c context.Context, connectionID RowID,
	format GetMessageFormat) ([]*Message, bool) {
	connection := csc.getConnection(c, connectionID)
	if connection.ID.IsZero() {
		return nil, false
	}

	messages := make([]*Message, 0, initialMessagesSize)
	var clientIndex, serverIndex uint64

	var clientBlocksIndex, serverBlocksIndex int
	var clientDocumentIndex, serverDocumentIndex int
	clientStream := csc.getConnectionStream(c, connectionID, true, clientDocumentIndex)
	serverStream := csc.getConnectionStream(c, connectionID, false, serverDocumentIndex)

	hasClientBlocks := func() bool {
		return clientBlocksIndex < len(clientStream.BlocksIndexes)
	}
	hasServerBlocks := func() bool {
		return serverBlocksIndex < len(serverStream.BlocksIndexes)
	}

	var message *Message
	messagesBuffer := make([]*Message, 0, 16)
	contentChunkBuffer := new(bytes.Buffer)
	var lastContentSlice []byte
	var sideChanged, lastClient, lastServer bool
	for !clientStream.ID.IsZero() || !serverStream.ID.IsZero() {
		if hasClientBlocks() && (!hasServerBlocks() || // next payload is from client
			clientStream.BlocksTimestamps[clientBlocksIndex].UnixNano() <=
				serverStream.BlocksTimestamps[serverBlocksIndex].UnixNano()) {
			start := clientStream.BlocksIndexes[clientBlocksIndex]
			end := 0
			if clientBlocksIndex < len(clientStream.BlocksIndexes)-1 {
				end = clientStream.BlocksIndexes[clientBlocksIndex+1]
			} else {
				end = len(clientStream.Payload)
			}
			size := uint64(end - start)

			message = &Message{
				FromClient:      true,
				Content:         DecodeBytes(clientStream.Payload[start:end], format.Format),
				Index:           start,
				Timestamp:       clientStream.BlocksTimestamps[clientBlocksIndex],
				IsRetransmitted: clientStream.BlocksLoss[clientBlocksIndex],
				RegexMatches:    findMatchesBetween(clientStream.PatternMatches, clientIndex, clientIndex+size),
			}
			clientIndex += size
			clientBlocksIndex++

			lastContentSlice = clientStream.Payload[start:end]
			sideChanged, lastClient, lastServer = lastServer, true, false
		} else { // next payload is from server
			start := serverStream.BlocksIndexes[serverBlocksIndex]
			end := 0
			if serverBlocksIndex < len(serverStream.BlocksIndexes)-1 {
				end = serverStream.BlocksIndexes[serverBlocksIndex+1]
			} else {
				end = len(serverStream.Payload)
			}
			size := uint64(end - start)

			message = &Message{
				FromClient:      false,
				Content:         DecodeBytes(serverStream.Payload[start:end], format.Format),
				Index:           start,
				Timestamp:       serverStream.BlocksTimestamps[serverBlocksIndex],
				IsRetransmitted: serverStream.BlocksLoss[serverBlocksIndex],
				RegexMatches:    findMatchesBetween(serverStream.PatternMatches, serverIndex, serverIndex+size),
			}
			serverIndex += size
			serverBlocksIndex++

			lastContentSlice = serverStream.Payload[start:end]
			sideChanged, lastClient, lastServer = lastClient, false, true
		}

		if !hasClientBlocks() {
			clientDocumentIndex++
			clientBlocksIndex = 0
			clientIndex = 0
			clientStream = csc.getConnectionStream(c, connectionID, true, clientDocumentIndex)
		}
		if !hasServerBlocks() {
			serverDocumentIndex++
			serverBlocksIndex = 0
			serverIndex = 0
			serverStream = csc.getConnectionStream(c, connectionID, false, serverDocumentIndex)
		}

		updateMetadata := func() {
			metadata := parsers.Parse(contentChunkBuffer.Bytes())
			var isMetadataContinuation bool
			for _, elem := range messagesBuffer {
				elem.Metadata = metadata
				elem.IsMetadataContinuation = metadata != nil && isMetadataContinuation
				isMetadataContinuation = true
			}

			messagesBuffer = messagesBuffer[:0]
			contentChunkBuffer.Reset()
		}

		if sideChanged {
			updateMetadata()
		}
		messagesBuffer = append(messagesBuffer, message)
		contentChunkBuffer.Write(lastContentSlice)

		if clientStream.ID.IsZero() && serverStream.ID.IsZero() {
			updateMetadata()
		}

		messages = append(messages, message)
	}

	return messages, true
}

func (csc ConnectionStreamsController) DownloadConnectionMessages(c context.Context, connectionID RowID,
	format DownloadMessageFormat) (string, bool) {
	connection := csc.getConnection(c, connectionID)
	if connection.ID.IsZero() {
		return "", false
	}

	var sb strings.Builder
	includeClient, includeServer := format.Type != "only_server", format.Type != "only_client"
	isPwntools := format.Type == "pwntools"

	var clientBlocksIndex, serverBlocksIndex int
	var clientDocumentIndex, serverDocumentIndex int
	var clientStream ConnectionStream
	if includeClient {
		clientStream = csc.getConnectionStream(c, connectionID, true, clientDocumentIndex)
	}
	var serverStream ConnectionStream
	if includeServer {
		serverStream = csc.getConnectionStream(c, connectionID, false, serverDocumentIndex)
	}

	hasClientBlocks := func() bool {
		return clientBlocksIndex < len(clientStream.BlocksIndexes)
	}
	hasServerBlocks := func() bool {
		return serverBlocksIndex < len(serverStream.BlocksIndexes)
	}

	if isPwntools {
		if format.Format == "base32" || format.Format == "base64" {
			sb.WriteString("import base64\n")
		}
		sb.WriteString("from pwn import *\n\n")
		sb.WriteString(fmt.Sprintf("p = remote('%s', %d)\n", connection.DestinationIP, connection.DestinationPort))
	}

	lastIsClient, lastIsServer := true, true
	for !clientStream.ID.IsZero() || !serverStream.ID.IsZero() {
		if hasClientBlocks() && (!hasServerBlocks() || // next payload is from client
			clientStream.BlocksTimestamps[clientBlocksIndex].UnixNano() <=
				serverStream.BlocksTimestamps[serverBlocksIndex].UnixNano()) {
			start := clientStream.BlocksIndexes[clientBlocksIndex]
			end := 0
			if clientBlocksIndex < len(clientStream.BlocksIndexes)-1 {
				end = clientStream.BlocksIndexes[clientBlocksIndex+1]
			} else {
				end = len(clientStream.Payload)
			}

			if !lastIsClient {
				sb.WriteString("\n")
			}
			lastIsClient = true
			lastIsServer = false
			if isPwntools {
				sb.WriteString(decodePwntools(clientStream.Payload[start:end], true, format.Format))
			} else {
				sb.WriteString(DecodeBytes(clientStream.Payload[start:end], format.Format))
			}
			clientBlocksIndex++
		} else { // next payload is from server
			start := serverStream.BlocksIndexes[serverBlocksIndex]
			end := 0
			if serverBlocksIndex < len(serverStream.BlocksIndexes)-1 {
				end = serverStream.BlocksIndexes[serverBlocksIndex+1]
			} else {
				end = len(serverStream.Payload)
			}

			if !lastIsServer {
				sb.WriteString("\n")
			}
			lastIsClient = false
			lastIsServer = true
			if isPwntools {
				sb.WriteString(decodePwntools(serverStream.Payload[start:end], false, format.Format))
			} else {
				sb.WriteString(DecodeBytes(serverStream.Payload[start:end], format.Format))
			}
			serverBlocksIndex++
		}

		if includeClient && !hasClientBlocks() {
			clientDocumentIndex++
			clientBlocksIndex = 0
			clientStream = csc.getConnectionStream(c, connectionID, true, clientDocumentIndex)
		}
		if includeServer && !hasServerBlocks() {
			serverDocumentIndex++
			serverBlocksIndex = 0
			serverStream = csc.getConnectionStream(c, connectionID, false, serverDocumentIndex)
		}
	}

	return sb.String(), true
}

func (csc ConnectionStreamsController) getConnection(c context.Context, connectionID RowID) Connection {
	var connection Connection
	if err := csc.storage.Find(Connections).Context(c).Filter(OrderedDocument{{"_id", connectionID}}).
		First(&connection); err != nil {
		log.WithError(err).WithField("id", connectionID).Panic("failed to get connection")
	}
	return connection
}

func (csc ConnectionStreamsController) getConnectionStream(c context.Context, connectionID RowID, fromClient bool,
	documentIndex int) ConnectionStream {
	var result ConnectionStream
	if err := csc.storage.Find(ConnectionStreams).Filter(OrderedDocument{
		{"connection_id", connectionID},
		{"from_client", fromClient},
		{"document_index", documentIndex},
	}).Context(c).First(&result); err != nil {
		log.WithError(err).WithField("connection_id", connectionID).Panic("failed to get a ConnectionStream")
	}
	return result
}

func findMatchesBetween(patternMatches map[uint][]PatternSlice, from, to uint64) []RegexSlice {
	regexSlices := make([]RegexSlice, 0, initialRegexSlicesCount)
	for _, slices := range patternMatches {
		for _, slice := range slices {
			if from > slice[1] || to <= slice[0] {
				continue
			}

			var start, end uint64
			if from > slice[0] {
				start = 0
			} else {
				start = slice[0] - from
			}

			if to <= slice[1] {
				end = to - from
			} else {
				end = slice[1] - from
			}

			regexSlices = append(regexSlices, RegexSlice{From: start, To: end})
		}
	}
	return regexSlices
}

func decodePwntools(payload []byte, isClient bool, format string) string {
	if !isClient && len(payload) > pwntoolsMaxServerBytes {
		payload = payload[len(payload)-pwntoolsMaxServerBytes:]
	}

	var content string
	switch format {
	case "hex":
		content = fmt.Sprintf("bytes.fromhex('%s')", DecodeBytes(payload, format))
	case "base32":
		content = fmt.Sprintf("base64.b32decode('%s')", DecodeBytes(payload, format))
	case "base64":
		content = fmt.Sprintf("base64.b64decode('%s')", DecodeBytes(payload, format))
	default:
		content = fmt.Sprintf("'%s'", strings.Replace(DecodeBytes(payload, "ascii"), "'", "\\'", -1))
	}

	if isClient {
		return fmt.Sprintf("p.send(%s)\n", content)
	}

	return fmt.Sprintf("p.recvuntil(%s)\n", content)
}
