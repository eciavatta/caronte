package main

import (
	"context"
	"errors"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"github.com/google/gopacket/tcpassembly"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	"net"
	"strconv"
	"sync"
	"time"
)

const initialAssemblerPoolSize = 16
const flushOlderThan = 5 * time.Minute
const invalidSessionId = "invalid_id"
const importUpdateProgressInterval = 3 * time.Second
const initialPacketPerServicesMapSize = 16
const importedPcapsCollectionName = "imported_pcaps"


type PcapImporter struct {
	storage     Storage
	streamPool  *tcpassembly.StreamPool
	assemblers  []*tcpassembly.Assembler
	sessions    map[string]context.CancelFunc
	mAssemblers sync.Mutex
	mSessions   sync.Mutex
	serverIp    gopacket.Endpoint
}

type flowCount [2]int


func NewPcapImporter(storage Storage, serverIp net.IP) *PcapImporter {
	serverEndpoint := layers.NewIPEndpoint(serverIp)
	streamFactory := &BiDirectionalStreamFactory{
		storage: storage,
		serverIp: serverEndpoint,
	}
	streamPool := tcpassembly.NewStreamPool(streamFactory)

	return &PcapImporter{
		storage:     storage,
		streamPool:  streamPool,
		assemblers:  make([]*tcpassembly.Assembler, 0, initialAssemblerPoolSize),
		sessions:    make(map[string]context.CancelFunc),
		mAssemblers: sync.Mutex{},
		mSessions:   sync.Mutex{},
		serverIp:    serverEndpoint,
	}
}

// Import a pcap file to the database. The pcap file must be present at the fileName path. If the pcap is already
// going to be imported or if it has been already imported in the past the function returns an error. Otherwise it
// create a new session and starts to import the pcap, and returns immediately the session name (that is the sha256
// of the pcap).
func (pi *PcapImporter) ImportPcap(fileName string) (string, error) {
	hash, err := Sha256Sum(fileName)
	if err != nil {
		return invalidSessionId, err
	}

	pi.mSessions.Lock()
	_, ok := pi.sessions[hash]
	if ok {
		pi.mSessions.Unlock()
		return hash, errors.New("another equal session in progress")
	}

	doc := OrderedDocument{
		{"_id", hash},
		{"started_at", time.Now()},
		{"completed_at", nil},
		{"processed_packets", 0},
		{"invalid_packets", 0},
		{"packets_per_services", nil},
		{"importing_error", err},
	}
	ctx, canc := context.WithCancel(context.Background())
	_, err = pi.storage.InsertOne(ctx, importedPcapsCollectionName, doc)
	if err != nil {
		pi.mSessions.Unlock()
		_, alreadyProcessed := err.(mongo.WriteException)
		if alreadyProcessed {
			return hash, errors.New("pcap already processed")
		}
		return hash, err
	}
	pi.sessions[hash] = canc
	pi.mSessions.Unlock()

	go pi.parsePcap(hash, fileName, ctx)

	return hash, nil
}

func (pi *PcapImporter) CancelImport(sessionId string) error {
	pi.mSessions.Lock()
	defer pi.mSessions.Unlock()
	cancel, ok := pi.sessions[sessionId]
	if ok {
		delete(pi.sessions, sessionId)
		cancel()
		return nil
	} else {
		return errors.New("session " + sessionId + " not found")
	}
}

// Read the pcap and save the tcp stream flow to the database
func (pi *PcapImporter) parsePcap(sessionId, fileName string, ctx context.Context) {
	handle, err := pcap.OpenOffline(fileName)
	if err != nil {
		// TODO: update db and set error
		return
	}

	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
	packetSource.NoCopy = true
	assembler := pi.takeAssembler()
	packets := packetSource.Packets()
	firstPacketTime := time.Time{}
	updateProgressInterval := time.Tick(importUpdateProgressInterval)

	processedPackets := 0
	invalidPackets := 0
	packetsPerService := make(map[int]*flowCount, initialPacketPerServicesMapSize)

	progressUpdate := func(completed bool, err error) {
		update := UnorderedDocument{
			"processed_packets": processedPackets,
			"invalid_packets": invalidPackets,
			"packets_per_services": packetsPerService,
			"importing_error": err,
		}
		if completed {
			update["completed_at"] = time.Now()
		}

		_, _err := pi.storage.UpdateOne(nil, importedPcapsCollectionName, OrderedDocument{{"_id", sessionId}},
			completed, false)

		if _err != nil {
			log.Println("can't update importing statistics : ", _err)
		}
	}

	deleteSession := func() {
		pi.mSessions.Lock()
		delete(pi.sessions, sessionId)
		pi.mSessions.Unlock()
	}

	for {
		select {
		case <- ctx.Done():
			handle.Close()
			deleteSession()
			progressUpdate(false,	errors.New("import process cancelled"))
			return
		default:
		}

		select {
		case packet := <-packets:
			if packet == nil { // completed
				if !firstPacketTime.IsZero() {
					assembler.FlushOlderThan(firstPacketTime.Add(-flushOlderThan))
				}
				pi.releaseAssembler(assembler)
				handle.Close()

				deleteSession()
				progressUpdate(true, nil)

				return
			}
			processedPackets++

			if packet.NetworkLayer() == nil || packet.TransportLayer() == nil ||
				packet.TransportLayer().LayerType() != layers.LayerTypeTCP { // invalid packet
				invalidPackets++
				continue
			}

			timestamp := packet.Metadata().Timestamp
			if firstPacketTime.IsZero() {
				firstPacketTime = timestamp
			}

			tcp := packet.TransportLayer().(*layers.TCP)
			var servicePort, index int
			if packet.NetworkLayer().NetworkFlow().Dst() == pi.serverIp {
				servicePort, _ = strconv.Atoi(tcp.DstPort.String())
				index = 0
			} else {
				servicePort, _ = strconv.Atoi(tcp.SrcPort.String())
				index = 1
			}
			fCount, ok := packetsPerService[servicePort]
			if !ok {
				fCount = &flowCount{0, 0}
				packetsPerService[servicePort] = fCount
			}
			fCount[index]++

			assembler.AssembleWithTimestamp(packet.NetworkLayer().NetworkFlow(), tcp, timestamp)
		case <-updateProgressInterval:
			progressUpdate(false, nil)
		}
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
