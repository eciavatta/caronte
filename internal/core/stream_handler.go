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

package core

import (
	"bytes"
	"io"
	"strings"
	"time"

	"github.com/eciavatta/caronte/internal/similarity"
	"github.com/eciavatta/caronte/pkg/tcpassembly"
	"github.com/flier/gohs/hyperscan"
	"github.com/glaslos/tlsh"
	"github.com/google/gopacket"
	log "github.com/sirupsen/logrus"
)

const MaxDocumentSize = 1024 * 1024
const InitialBlockCount = 1024
const InitialPatternSliceSize = 8
const InitialPacketsCount = 16
const MaxPacketsCount = 1024

// IMPORTANT:  If you use a StreamHandler, you MUST read ALL BYTES from it,
// quickly.  Not reading available bytes will block TCP stream reassembly.  It's
// a common pattern to do this by starting a goroutine in the factory's New
// method:
type StreamHandler struct {
	connection        ConnectionHandler
	streamFlow        StreamFlow
	buffer            *bytes.Buffer
	indexes           []int
	timestamps        []time.Time
	lossBlocks        []bool
	currentIndex      int
	firstPacketSeen   time.Time
	lastPacketSeen    time.Time
	documentsIDs      []RowID
	streamLength      int
	patternStream     hyperscan.Stream
	patternMatches    map[uint][]PatternSlice
	scanner           Scanner
	isClient          bool
	tlshHash          string
	byteHistogramHash []byte
	packets           []gopacket.Packet
}

// NewReaderStream returns a new StreamHandler object.
func NewStreamHandler(connection ConnectionHandler, streamFlow StreamFlow, scanner Scanner, isClient bool) StreamHandler {
	handler := StreamHandler{
		connection:     connection,
		streamFlow:     streamFlow,
		buffer:         new(bytes.Buffer),
		indexes:        make([]int, 0, InitialBlockCount),
		timestamps:     make([]time.Time, 0, InitialBlockCount),
		lossBlocks:     make([]bool, 0, InitialBlockCount),
		documentsIDs:   make([]RowID, 0, 1), // most of the time the stream fit in one document
		patternMatches: make(map[uint][]PatternSlice, connection.PatternsDatabaseSize()),
		scanner:        scanner,
		isClient:       isClient,
		packets:        make([]gopacket.Packet, 0, InitialPacketsCount),
	}

	stream, err := connection.PatternsDatabase().Open(0, scanner.scratch, handler.onMatch, nil)
	if err != nil {
		log.WithField("streamFlow", streamFlow).WithError(err).Error("failed to create a stream")
	}
	handler.patternStream = stream

	return handler
}

// Reassembled implements tcpassembly.Stream's Reassembled function.
func (sh *StreamHandler) Reassembled(reassembly []tcpassembly.Reassembly) {
	for _, r := range reassembly {
		skip := r.Skip
		isLoss := skip != 0

		if r.Start {
			sh.firstPacketSeen = r.Seen
		}
		if r.End {
			sh.lastPacketSeen = r.Seen
		}

		reassemblyLen := len(r.Bytes)
		if reassemblyLen == 0 {
			continue
		}
		if skip < 0 || skip >= reassemblyLen { // start or flush ~ workaround
			skip = 0
		}

		if sh.buffer.Len()+len(r.Bytes)-skip > MaxDocumentSize*(len(sh.documentsIDs)+1) {
			sh.storageCurrentDocument()
			sh.resetCurrentDocument()
		}
		n, err := sh.buffer.Write(r.Bytes[skip:])
		if err != nil {
			log.WithError(err).Error("failed to copy bytes from a Reassemble")
			continue
		}
		sh.indexes = append(sh.indexes, sh.currentIndex)
		sh.timestamps = append(sh.timestamps, r.Seen)
		sh.lossBlocks = append(sh.lossBlocks, isLoss)
		sh.currentIndex += n
		sh.streamLength += n

		if sh.patternStream != nil {
			err = sh.patternStream.Scan(r.Bytes[skip:])
			if err != nil {
				log.WithError(err).Error("failed to scan packet buffer")
			}
		}
	}
}

// ReassemblyComplete implements tcpassembly.Stream's ReassemblyComplete function.
func (sh *StreamHandler) ReassemblyComplete() {
	if sh.patternStream != nil {
		err := sh.patternStream.Close()
		if err != nil {
			log.WithError(err).Error("failed to close pattern stream")
		}
	}

	if sh.currentIndex > 0 {
		sh.storageCurrentDocument()
	}

	if tlshHash, err := tlsh.HashBytes(sh.buffer.Bytes()); err == nil {
		sh.tlshHash = tlshHash.String()
	} else if err != io.EOF {
		log.WithError(err).Error("failed to hash stream with tlsh")
	}

	sh.byteHistogramHash = similarity.ByteHistogramDigest(sh.buffer.Bytes()).Digest()

	sh.connection.Complete(sh)
}

// ReassemblyComplete implements tcpassembly.Stream's Packet function.
func (sh *StreamHandler) Packet(packet gopacket.Packet) {
	if len(sh.packets) >= MaxPacketsCount {
		return
	}

	sh.packets = append(sh.packets, packet)
}

func (sh *StreamHandler) resetCurrentDocument() {
	sh.indexes = sh.indexes[:0]
	sh.timestamps = sh.timestamps[:0]
	sh.lossBlocks = sh.lossBlocks[:0]
	sh.currentIndex = 0

	for i := range sh.patternMatches {
		sh.patternMatches[i] = sh.patternMatches[i][:0]
	}
}

func (sh *StreamHandler) onMatch(id uint, from uint64, to uint64, _ uint, _ any) error {
	patternSlices, isPresent := sh.patternMatches[id]
	if isPresent {
		if len(patternSlices) > 0 {
			lastElement := &patternSlices[len(patternSlices)-1]
			if lastElement[0] == from { // make the regex greedy to match the maximum number of chars
				lastElement[1] = to
				return nil
			}
		}
		// new from == new match
		sh.patternMatches[id] = append(patternSlices, PatternSlice{from, to})
	} else {
		patternSlices = make([]PatternSlice, 1, InitialPatternSliceSize)
		patternSlices[0] = PatternSlice{from, to}
		sh.patternMatches[id] = patternSlices
	}

	return nil
}

func (sh *StreamHandler) storageCurrentDocument() {
	payload := sh.streamFlow.Hash()&uint64(0xffffffffffffff00) | uint64(len(sh.documentsIDs)) // LOL
	streamID := CustomRowID(payload, sh.firstPacketSeen)

	bufferStart := len(sh.documentsIDs) * MaxDocumentSize
	bufferEnd := bufferStart + sh.currentIndex
	if _, err := sh.connection.Storage().Insert(ConnectionStreams).
		One(ConnectionStream{
			ID:               streamID,
			ConnectionID:     ZeroRowID,
			DocumentIndex:    len(sh.documentsIDs),
			Payload:          sh.buffer.Bytes()[bufferStart:bufferEnd],
			PayloadString:    strings.ToValidUTF8(string(sh.buffer.Bytes()[bufferStart:bufferEnd]), ""),
			BlocksIndexes:    sh.indexes,
			BlocksTimestamps: sh.timestamps,
			BlocksLoss:       sh.lossBlocks,
			PatternMatches:   sh.patternMatches,
			FromClient:       sh.isClient,
		}); err != nil {
		log.WithError(err).Error("failed to insert connection stream")
	} else {
		sh.documentsIDs = append(sh.documentsIDs, streamID)
	}
}
