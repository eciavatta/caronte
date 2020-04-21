package main

import (
	"context"
	"errors"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"github.com/google/gopacket/tcpassembly"
	log "github.com/sirupsen/logrus"
	"net"
	"sync"
	"time"
)

const initialAssemblerPoolSize = 16
const flushOlderThan = 5 * time.Minute
const importUpdateProgressInterval = 100 * time.Millisecond

type PcapImporter struct {
	storage     Storage
	streamPool  *tcpassembly.StreamPool
	assemblers  []*tcpassembly.Assembler
	sessions    map[string]ImportingSession
	mAssemblers sync.Mutex
	mSessions   sync.Mutex
	serverIP    gopacket.Endpoint
}

type ImportingSession struct {
	ID                string               `json:"id" bson:"_id"`
	StartedAt         time.Time            `json:"started_at" bson:"started_at"`
	Size              int64                `json:"size" bson:"size"`
	CompletedAt       time.Time            `json:"completed_at" bson:"completed_at,omitempty"`
	ProcessedPackets  int                  `json:"processed_packets" bson:"processed_packets"`
	InvalidPackets    int                  `json:"invalid_packets" bson:"invalid_packets"`
	PacketsPerService map[uint16]flowCount `json:"packets_per_service" bson:"packets_per_service"`
	ImportingError    string               `json:"importing_error" bson:"importing_error,omitempty"`
	cancelFunc        context.CancelFunc
	completed         chan string
}

type flowCount [2]int

func NewPcapImporter(storage Storage, serverIP net.IP, rulesManager RulesManager) *PcapImporter {
	serverEndpoint := layers.NewIPEndpoint(serverIP)
	streamPool := tcpassembly.NewStreamPool(NewBiDirectionalStreamFactory(storage, serverEndpoint, rulesManager))

	return &PcapImporter{
		storage:     storage,
		streamPool:  streamPool,
		assemblers:  make([]*tcpassembly.Assembler, 0, initialAssemblerPoolSize),
		sessions:    make(map[string]ImportingSession),
		mAssemblers: sync.Mutex{},
		mSessions:   sync.Mutex{},
		serverIP:    serverEndpoint,
	}
}

// Import a pcap file to the database. The pcap file must be present at the fileName path. If the pcap is already
// going to be imported or if it has been already imported in the past the function returns an error. Otherwise it
// create a new session and starts to import the pcap, and returns immediately the session name (that is the sha256
// of the pcap).
func (pi *PcapImporter) ImportPcap(fileName string) (string, error) {
	hash, err := Sha256Sum(fileName)
	if err != nil {
		return "", err
	}

	pi.mSessions.Lock()
	_, isPresent := pi.sessions[hash]
	if isPresent {
		pi.mSessions.Unlock()
		return hash, errors.New("pcap already processed")
	}

	ctx, cancelFunc := context.WithCancel(context.Background())
	session := ImportingSession{
		ID:                hash,
		StartedAt:         time.Now(),
		Size:              FileSize(fileName),
		PacketsPerService: make(map[uint16]flowCount),
		cancelFunc:        cancelFunc,
		completed:         make(chan string),
	}

	pi.sessions[hash] = session
	pi.mSessions.Unlock()

	go pi.parsePcap(session, fileName, ctx)

	return hash, nil
}

func (pi *PcapImporter) GetSession(sessionID string) (ImportingSession, bool) {
	pi.mSessions.Lock()
	defer pi.mSessions.Unlock()
	session, isPresent := pi.sessions[sessionID]
	return session, isPresent
}

func (pi *PcapImporter) CancelSession(sessionID string) error {
	pi.mSessions.Lock()
	defer pi.mSessions.Unlock()
	if session, isPresent := pi.sessions[sessionID]; !isPresent {
		return errors.New("session " + sessionID + " not found")
	} else {
		session.cancelFunc()
		return nil
	}
}

// Read the pcap and save the tcp stream flow to the database
func (pi *PcapImporter) parsePcap(session ImportingSession, fileName string, ctx context.Context) {
	handle, err := pcap.OpenOffline(fileName)
	if err != nil {
		pi.progressUpdate(session, false, "failed to process pcap")
		log.WithError(err).WithFields(log.Fields{"session": session, "fileName": fileName}).
			Error("failed to open pcap")
		return
	}

	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
	packetSource.NoCopy = true
	assembler := pi.takeAssembler()
	packets := packetSource.Packets()
	firstPacketTime := time.Time{}
	updateProgressInterval := time.Tick(importUpdateProgressInterval)

	for {
		select {
		case <-ctx.Done():
			handle.Close()
			pi.releaseAssembler(assembler)
			pi.progressUpdate(session, false, "import process cancelled")
			return
		default:
		}

		select {
		case packet := <-packets:
			if packet == nil { // completed
				if !firstPacketTime.IsZero() {
					assembler.FlushOlderThan(firstPacketTime.Add(-flushOlderThan))
				}
				handle.Close()
				pi.releaseAssembler(assembler)
				pi.progressUpdate(session, true, "")
				return
			}

			timestamp := packet.Metadata().Timestamp
			if firstPacketTime.IsZero() {
				firstPacketTime = timestamp
			}

			session.ProcessedPackets++

			if packet.NetworkLayer() == nil || packet.TransportLayer() == nil ||
				packet.TransportLayer().LayerType() != layers.LayerTypeTCP { // invalid packet
				session.InvalidPackets++
				continue
			}

			tcp := packet.TransportLayer().(*layers.TCP)
			var servicePort uint16
			var index int
			isDstServer := packet.NetworkLayer().NetworkFlow().Dst() == pi.serverIP
			isSrcServer := packet.NetworkLayer().NetworkFlow().Src() == pi.serverIP
			if isDstServer && !isSrcServer {
				servicePort = uint16(tcp.DstPort)
				index = 0
			} else if isSrcServer && !isDstServer {
				servicePort = uint16(tcp.SrcPort)
				index = 1
			} else {
				session.InvalidPackets++
				continue
			}
			fCount, isPresent := session.PacketsPerService[servicePort]
			if !isPresent {
				fCount = flowCount{0, 0}
			}
			fCount[index]++
			session.PacketsPerService[servicePort] = fCount

			assembler.AssembleWithTimestamp(packet.NetworkLayer().NetworkFlow(), tcp, timestamp)
		case <-updateProgressInterval:
			pi.progressUpdate(session, false, "")
		}
	}
}

func (pi *PcapImporter) progressUpdate(session ImportingSession, completed bool, err string) {
	if completed {
		session.CompletedAt = time.Now()
	}
	session.ImportingError = err

	packetsPerService := session.PacketsPerService
	session.PacketsPerService = make(map[uint16]flowCount, len(packetsPerService))
	for key, value := range packetsPerService {
		session.PacketsPerService[key] = value
	}

	pi.mSessions.Lock()
	pi.sessions[session.ID] = session
	pi.mSessions.Unlock()

	if completed || session.ImportingError != "" {
		if _, _err := pi.storage.Insert(ImportingSessions).One(session); _err != nil {
			log.WithError(_err).WithField("session", session).Error("failed to insert importing stats")
		}
		session.completed <- session.ImportingError
	}
}

func (pi *PcapImporter) takeAssembler() *tcpassembly.Assembler {
	pi.mAssemblers.Lock()
	defer pi.mAssemblers.Unlock()

	if len(pi.assemblers) == 0 {
		return tcpassembly.NewAssembler(pi.streamPool)
	}

	index := len(pi.assemblers) - 1
	assembler := pi.assemblers[index]
	pi.assemblers = pi.assemblers[:index]

	return assembler
}

func (pi *PcapImporter) releaseAssembler(assembler *tcpassembly.Assembler) {
	pi.mAssemblers.Lock()
	pi.assemblers = append(pi.assemblers, assembler)
	pi.mAssemblers.Unlock()
}
