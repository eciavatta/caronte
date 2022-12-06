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
	"encoding/binary"
	"fmt"
	"hash/fnv"
	"net"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/eciavatta/caronte/pkg/tcpassembly"
	"github.com/flier/gohs/hyperscan"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcapgo"
	log "github.com/sirupsen/logrus"
)

const initialConnectionsCapacity = 1024
const initialScannersCapacity = 1024

type BiDirectionalStreamFactory struct {
	storage                 Storage
	serverNet               *net.IPNet
	connections             map[StreamFlow]ConnectionHandler
	mConnections            sync.Mutex
	rulesManager            RulesManager
	rulesDatabase           RulesDatabase
	mRulesDatabase          sync.Mutex
	scanners                []Scanner
	connectionStatusChannel chan bool // true if completed, false if pending
	notificationController  NotificationController
}

type StreamFlow [4]gopacket.Endpoint

type Scanner struct {
	scratch *hyperscan.Scratch
	version RowID
}

type ConnectionsStatistics struct {
	PendingConnections   uint `json:"pending_connections"`
	CompletedConnections uint `json:"completed_connections"`
	ConnectionsPerMinute uint `json:"connections_per_minute"`
}

type ConnectionHandler interface {
	Complete(handler *StreamHandler)
	Storage() Storage
	PatternsDatabase() hyperscan.StreamDatabase
	PatternsDatabaseSize() int
}

type connectionHandlerImpl struct {
	factory        *BiDirectionalStreamFactory
	connectionFlow StreamFlow
	mComplete      sync.Mutex
	otherStream    *StreamHandler
	linkType       layers.LinkType
}

func NewBiDirectionalStreamFactory(storage Storage, serverNet *net.IPNet,
	rulesManager RulesManager, notificationController NotificationController) *BiDirectionalStreamFactory {

	factory := &BiDirectionalStreamFactory{
		storage:                 storage,
		serverNet:               serverNet,
		connections:             make(map[StreamFlow]ConnectionHandler, initialConnectionsCapacity),
		mConnections:            sync.Mutex{},
		rulesManager:            rulesManager,
		mRulesDatabase:          sync.Mutex{},
		scanners:                make([]Scanner, 0, initialScannersCapacity),
		connectionStatusChannel: make(chan bool, 4096),
		notificationController:  notificationController,
	}

	go factory.updateRulesDatabaseService()
	go factory.notificationService()
	return factory
}

func (factory *BiDirectionalStreamFactory) updateRulesDatabaseService() {
	for {
		rulesDatabase, ok := <-factory.rulesManager.DatabaseUpdateChannel()
		if !ok {
			return
		}
		factory.mRulesDatabase.Lock()
		scanners := factory.scanners
		factory.scanners = factory.scanners[:0]

		for _, s := range scanners {
			err := s.scratch.Realloc(rulesDatabase.database)
			if err != nil {
				log.WithError(err).Error("failed to realloc an existing scanner")
			} else {
				s.version = rulesDatabase.version
				factory.scanners = append(factory.scanners, s)
			}
		}

		factory.rulesDatabase = rulesDatabase
		factory.mRulesDatabase.Unlock()
	}
}

func (factory *BiDirectionalStreamFactory) takeScanner() Scanner {
	factory.mRulesDatabase.Lock()
	defer factory.mRulesDatabase.Unlock()

	if len(factory.scanners) == 0 {
		scratch, err := hyperscan.NewScratch(factory.rulesDatabase.database)
		if err != nil {
			log.WithError(err).Fatal("failed to alloc a new scratch")
		}

		return Scanner{
			scratch: scratch,
			version: factory.rulesDatabase.version,
		}
	}

	index := len(factory.scanners) - 1
	scanner := factory.scanners[index]
	factory.scanners = factory.scanners[:index]

	return scanner
}

func (factory *BiDirectionalStreamFactory) releaseScanner(scanner Scanner) {
	factory.mRulesDatabase.Lock()
	defer factory.mRulesDatabase.Unlock()

	if scanner.version != factory.rulesDatabase.version {
		err := scanner.scratch.Realloc(factory.rulesDatabase.database)
		if err != nil {
			log.WithError(err).Error("failed to realloc an existing scanner")
			return
		}
		scanner.version = factory.rulesDatabase.version
	}
	factory.scanners = append(factory.scanners, scanner)
}

func (factory *BiDirectionalStreamFactory) New(netFlow, transportFlow gopacket.Flow, isServer bool, linkType layers.LinkType) tcpassembly.Stream {
	flow := StreamFlow{netFlow.Src(), netFlow.Dst(), transportFlow.Src(), transportFlow.Dst()}
	invertedFlow := StreamFlow{netFlow.Dst(), netFlow.Src(), transportFlow.Dst(), transportFlow.Src()}

	factory.mConnections.Lock()
	connection, isPresent := factory.connections[invertedFlow]

	if factory.serverNet != nil {
		isServer = factory.serverNet.Contains(netFlow.Src().Raw())
	}

	if isPresent {
		delete(factory.connections, invertedFlow)
	} else {
		var connectionFlow StreamFlow
		if isServer {
			connectionFlow = invertedFlow
		} else {
			connectionFlow = flow
		}
		connection = &connectionHandlerImpl{
			connectionFlow: connectionFlow,
			mComplete:      sync.Mutex{},
			factory:        factory,
		}
		factory.connections[flow] = connection
	}
	connection.(*connectionHandlerImpl).linkType = linkType // Ok I could have done better, I admit it
	factory.mConnections.Unlock()

	streamHandler := NewStreamHandler(connection, flow, factory.takeScanner(), !isServer)

	factory.connectionStatusChannel <- false

	return &streamHandler
}

func (ch *connectionHandlerImpl) Complete(handler *StreamHandler) {
	ch.factory.releaseScanner(handler.scanner)
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
	if handler.streamFlow == ch.connectionFlow {
		client = handler
		server = ch.otherStream
	} else {
		client = ch.otherStream
		server = handler
	}

	connectionID := CustomRowID(ch.connectionFlow.Hash(), startedAt)
	connection := Connection{
		ID:                      connectionID,
		SourceIP:                ch.connectionFlow[0].String(),
		DestinationIP:           ch.connectionFlow[1].String(),
		SourcePort:              binary.BigEndian.Uint16(ch.connectionFlow[2].Raw()),
		DestinationPort:         binary.BigEndian.Uint16(ch.connectionFlow[3].Raw()),
		StartedAt:               startedAt,
		ClosedAt:                closedAt,
		ClientBytes:             client.streamLength,
		ServerBytes:             server.streamLength,
		ClientDocuments:         len(client.documentsIDs),
		ServerDocuments:         len(server.documentsIDs),
		ProcessedAt:             time.Now(),
		ClientTlshHash:          client.tlshHash,
		ClientByteHistogramHash: client.byteHistogramHash,
		ServerTlshHash:          server.tlshHash,
		ServerByteHistogramHash: server.byteHistogramHash,
	}
	ch.factory.rulesManager.FillWithMatchedRules(&connection, client.patternMatches, server.patternMatches)

	_, err := ch.Storage().Insert(Connections).One(connection)
	if err != nil {
		log.WithError(err).WithField("connection", connection).Error("failed to insert a connection")
		return
	}

	streamsIDs := append(client.documentsIDs, server.documentsIDs...)
	if len(streamsIDs) > 0 {
		n, err := ch.Storage().Update(ConnectionStreams).
			Filter(OrderedDocument{{Key: "_id", Value: UnorderedDocument{"$in": streamsIDs}}}).
			Many(UnorderedDocument{"connection_id": connectionID})
		if err != nil {
			log.WithError(err).WithField("connection", connection).Error("failed to update connection streams")
		} else if int(n) != len(streamsIDs) {
			log.WithError(err).WithField("connection", connection).Error("failed to update all connections streams")
		}
	}

	ch.SaveConnectionPcap(connectionID.Hex(), client, server)
	ch.UpdateStatistics(connection)
	ch.factory.connectionStatusChannel <- true
}

func (ch *connectionHandlerImpl) SaveConnectionPcap(connectionID string, clientStream *StreamHandler, serverStream *StreamHandler) {
	c, s := 0, 0
	lc, ls := len(clientStream.packets), len(serverStream.packets)

	if lc == 0 && ls == 0 {
		return
	}

	f, err := os.Create(filepath.Join(ConnectionPcapsBasePath, fmt.Sprintf("%s.pcap", connectionID)))
	if err != nil {
		log.WithError(err).WithField("connectionID", connectionID).Error("failed to create connection pcap")
		return
	}
	defer f.Close()

	w := pcapgo.NewWriter(f)
	if err := w.WriteFileHeader(65536, ch.linkType); err != nil {
		log.WithError(err).WithField("connectionID", connectionID).Error("failed to write connection pcap header")
		return
	}

	writePacket := func(packet gopacket.Packet) error {
		return w.WritePacket(packet.Metadata().CaptureInfo, packet.Data())
	}

	for {
		if c < lc {
			if s < ls && serverStream.packets[s].Metadata().Timestamp.Before(clientStream.packets[c].Metadata().Timestamp) {
				err = writePacket(serverStream.packets[s])
				s++
			} else {
				err = writePacket(clientStream.packets[c])
				c++
			}
		} else if s < ls {
			err = writePacket(serverStream.packets[s])
			s++
		} else {
			return
		}

		if err != nil {
			log.WithError(err).WithField("connectionID", connectionID).Error("failed to write connection pcap packet")
			return
		}
	}
}

func (ch *connectionHandlerImpl) UpdateStatistics(connection Connection) {
	rangeStart := connection.StartedAt.Unix() / 60 // group statistic records by minutes
	duration := connection.ClosedAt.Sub(connection.StartedAt)
	// if one of the two parts doesn't close connection, the duration is +infinity or -infinity
	if duration.Hours() > 1 || duration.Hours() < -1 {
		duration = 0
	}
	servicePort := connection.DestinationPort

	updateDocument := UnorderedDocument{
		fmt.Sprintf("connections_per_service.%d", servicePort):  1,
		fmt.Sprintf("client_bytes_per_service.%d", servicePort): connection.ClientBytes,
		fmt.Sprintf("server_bytes_per_service.%d", servicePort): connection.ServerBytes,
		fmt.Sprintf("total_bytes_per_service.%d", servicePort):  connection.ClientBytes + connection.ServerBytes,
		fmt.Sprintf("duration_per_service.%d", servicePort):     duration.Milliseconds(),
	}

	for _, ruleID := range connection.MatchedRules {
		updateDocument[fmt.Sprintf("matched_rules.%s", ruleID.Hex())] = 1
	}

	var results interface{}
	if _, err := ch.Storage().Update(Statistics).Upsert(&results).
		Filter(OrderedDocument{{Key: "_id", Value: time.Unix(rangeStart*60, 0)}}).
		OneComplex(UnorderedDocument{"$inc": updateDocument}); err != nil {
		log.WithError(err).WithField("connection", connection).Error("failed to update connection statistics")
	}
}

func (factory *BiDirectionalStreamFactory) notificationService() {
	var stats, lastStats ConnectionsStatistics
	var connectionsPerMinute uint
	ticker := time.NewTicker(3 * time.Second)
	perMinuteTicker := time.NewTicker(time.Minute)

	updateStatistics := func() {
		lastStats = stats
		factory.notificationController.Notify("connections.statistics", stats)
	}

	for {
		select {
		case completed := <-factory.connectionStatusChannel:
			if completed {
				stats.CompletedConnections++
				stats.PendingConnections -= 2
				connectionsPerMinute++
			} else {
				stats.PendingConnections++
			}
		case <-ticker.C:
			if lastStats != stats {
				updateStatistics()
			}
		case <-perMinuteTicker.C:
			stats.ConnectionsPerMinute = connectionsPerMinute
			connectionsPerMinute = 0
		}
	}
}

func (ch *connectionHandlerImpl) Storage() Storage {
	return ch.factory.storage
}

func (ch *connectionHandlerImpl) PatternsDatabase() hyperscan.StreamDatabase {
	return ch.factory.rulesDatabase.database
}

func (ch *connectionHandlerImpl) PatternsDatabaseSize() int {
	return ch.factory.rulesDatabase.databaseSize
}

func (sf StreamFlow) Hash() uint64 {
	hash := fnv.New64a()
	_, _ = hash.Write(sf[0].Raw())
	_, _ = hash.Write(sf[1].Raw())
	_, _ = hash.Write(sf[2].Raw())
	_, _ = hash.Write(sf[3].Raw())
	return hash.Sum64()
}
