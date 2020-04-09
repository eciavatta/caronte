package main

import (
	"encoding/binary"
	"github.com/flier/gohs/hyperscan"
	"github.com/google/gopacket"
	"github.com/google/gopacket/tcpassembly"
	log "github.com/sirupsen/logrus"
	"sync"
	"time"
)

const initialConnectionsCapacity = 1024
const initialScannersCapacity = 1024

type BiDirectionalStreamFactory struct {
	storage        Storage
	serverIp       gopacket.Endpoint
	connections    map[StreamFlow]ConnectionHandler
	mConnections   sync.Mutex
	rulesManager   *RulesManager
	rulesDatabase  RulesDatabase
	mRulesDatabase sync.Mutex
	scanners       []Scanner
}

type StreamFlow [4]gopacket.Endpoint

type Scanner struct {
	scratch *hyperscan.Scratch
	version RowID
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
}

func NewBiDirectionalStreamFactory(storage Storage, serverIP gopacket.Endpoint,
	rulesManager *RulesManager) *BiDirectionalStreamFactory {

	factory := &BiDirectionalStreamFactory{
		storage:        storage,
		serverIp:       serverIP,
		connections:    make(map[StreamFlow]ConnectionHandler, initialConnectionsCapacity),
		mConnections:   sync.Mutex{},
		rulesManager:   rulesManager,
		mRulesDatabase: sync.Mutex{},
		scanners:       make([]Scanner, 0, initialScannersCapacity),
	}

	go factory.updateRulesDatabaseService()
	return factory
}

func (factory *BiDirectionalStreamFactory) updateRulesDatabaseService() {
	for {
		select {
		case rulesDatabase, ok := <-factory.rulesManager.databaseUpdated:
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
	}
	factory.scanners = append(factory.scanners, scanner)
}

func (factory *BiDirectionalStreamFactory) New(net, transport gopacket.Flow) tcpassembly.Stream {
	flow := StreamFlow{net.Src(), net.Dst(), transport.Src(), transport.Dst()}
	invertedFlow := StreamFlow{net.Dst(), net.Src(), transport.Dst(), transport.Src()}

	factory.mConnections.Lock()
	connection, isPresent := factory.connections[invertedFlow]
	if isPresent {
		delete(factory.connections, invertedFlow)
	} else {
		var connectionFlow StreamFlow
		if net.Src() == factory.serverIp {
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
	factory.mConnections.Unlock()

	streamHandler := NewStreamHandler(connection, flow, factory.takeScanner())

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

	connectionID := ch.Storage().NewCustomRowID(ch.connectionFlow.Hash(), startedAt)
	connection := Connection{
		ID:              connectionID,
		SourceIP:        ch.connectionFlow[0].String(),
		DestinationIP:   ch.connectionFlow[1].String(),
		SourcePort:      binary.BigEndian.Uint16(ch.connectionFlow[2].Raw()),
		DestinationPort: binary.BigEndian.Uint16(ch.connectionFlow[3].Raw()),
		StartedAt:       startedAt,
		ClosedAt:        closedAt,
		ClientPackets:   client.packetsCount,
		ServerPackets:   client.packetsCount,
		ClientBytes:     client.streamLength,
		ServerBytes:     server.streamLength,
		ClientDocuments: len(client.documentsIDs),
		ServerDocuments: len(server.documentsIDs),
		ProcessedAt:     time.Now(),
	}
	ch.factory.rulesManager.FillWithMatchedRules(&connection, client.patternMatches, server.patternMatches)

	_, err := ch.Storage().Insert(Connections).One(connection)
	if err != nil {
		log.WithError(err).WithField("connection", connection).Error("failed to insert a connection")
		return
	}

	streamsIds := append(client.documentsIDs, server.documentsIDs...)
	n, err := ch.Storage().Update(ConnectionStreams).
		Filter(OrderedDocument{{"_id", UnorderedDocument{"$in": streamsIds}}}).
		Many(UnorderedDocument{"connection_id": connectionID})
	if err != nil {
		log.WithError(err).WithField("connection", connection).Error("failed to update connection streams")
	} else if int(n) != len(streamsIds) {
		log.WithError(err).WithField("connection", connection).Error("failed to update all connections streams")
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
	return sf[0].FastHash() ^ sf[1].FastHash() ^ sf[2].FastHash() ^ sf[3].FastHash()
}
