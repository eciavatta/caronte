package main

import (
	"bufio"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/tcpassembly"
	"github.com/google/gopacket/tcpassembly/tcpreader"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net"
	"sync"
	"testing"
	"time"
)

func TestImportPcap(t *testing.T) {
	wrapper := NewTestStorageWrapper(t)
	pcapImporter := newTestPcapImporter(wrapper, "172.17.0.3")

	pcapImporter.releaseAssembler(pcapImporter.takeAssembler())

	sessionID, err := pcapImporter.ImportPcap("test_data/ping_pong_10000.pcap")
	require.NoError(t, err)

	duplicateSessionID, err := pcapImporter.ImportPcap("test_data/ping_pong_10000.pcap")
	require.Error(t, err)
	assert.Equal(t, sessionID, duplicateSessionID)

	_, isPresent := pcapImporter.GetSession("invalid")
	assert.False(t, isPresent)

	session := waitSessionCompletion(t, pcapImporter, sessionID)
	assert.Equal(t, 15008, session.ProcessedPackets)
	assert.Equal(t, 0, session.InvalidPackets)
	assert.Equal(t, map[uint16]flowCount{9999: {10004, 5004}}, session.PacketsPerService)
	assert.Zero(t, session.ImportingError)

	checkSessionEquals(t, wrapper, session)

	wrapper.Destroy(t)
}

func TestCancelImportSession(t *testing.T) {
	wrapper := NewTestStorageWrapper(t)
	pcapImporter := newTestPcapImporter(wrapper, "172.17.0.3")

	sessionID, err := pcapImporter.ImportPcap("test_data/ping_pong_10000.pcap")
	require.NoError(t, err)

	assert.Error(t, pcapImporter.CancelSession("invalid"))
	assert.NoError(t, pcapImporter.CancelSession(sessionID))

	session := waitSessionCompletion(t, pcapImporter, sessionID)
	assert.Zero(t, session.CompletedAt)
	assert.Equal(t, int64(1270696), session.Size)
	assert.Equal(t, 0, session.ProcessedPackets)
	assert.Equal(t, 0, session.InvalidPackets)
	assert.Equal(t, map[uint16]flowCount{}, session.PacketsPerService)
	assert.NotZero(t, session.ImportingError)

	checkSessionEquals(t, wrapper, session)

	wrapper.Destroy(t)
}

func TestImportNoTcpPackets(t *testing.T) {
	wrapper := NewTestStorageWrapper(t)
	pcapImporter := newTestPcapImporter(wrapper, "172.17.0.4")

	sessionID, err := pcapImporter.ImportPcap("test_data/icmp.pcap")
	require.NoError(t, err)

	session := waitSessionCompletion(t, pcapImporter, sessionID)
	assert.Equal(t, int64(228024), session.Size)
	assert.Equal(t, 2000, session.ProcessedPackets)
	assert.Equal(t, 2000, session.InvalidPackets)
	assert.Equal(t, map[uint16]flowCount{}, session.PacketsPerService)
	assert.Zero(t, session.ImportingError)

	checkSessionEquals(t, wrapper, session)

	wrapper.Destroy(t)
}

func newTestPcapImporter(wrapper *TestStorageWrapper, serverIP string) *PcapImporter {
	wrapper.AddCollection(ImportingSessions)

	serverEndpoint := layers.NewIPEndpoint(net.ParseIP(serverIP))
	streamPool := tcpassembly.NewStreamPool(&testStreamFactory{})

	return &PcapImporter{
		storage:     wrapper.Storage,
		streamPool:  streamPool,
		assemblers:  make([]*tcpassembly.Assembler, 0, initialAssemblerPoolSize),
		sessions:    make(map[string]ImportingSession),
		mAssemblers: sync.Mutex{},
		mSessions:   sync.Mutex{},
		serverIP:    serverEndpoint,
	}
}

func waitSessionCompletion(t *testing.T, pcapImporter *PcapImporter, sessionID string) ImportingSession {
	session, isPresent := pcapImporter.GetSession(sessionID)
	require.True(t, isPresent)
	<-session.completed

	session, isPresent = pcapImporter.GetSession(sessionID)
	assert.True(t, isPresent)
	assert.Equal(t, sessionID, session.ID)

	return session
}

func checkSessionEquals(t *testing.T, wrapper *TestStorageWrapper, session ImportingSession) {
	var result ImportingSession
	assert.NoError(t, wrapper.Storage.Find(ImportingSessions).Filter(OrderedDocument{{"_id", session.ID}}).
		Context(wrapper.Context).First(&result))
	assert.Equal(t, session.StartedAt.Unix(), result.StartedAt.Unix())
	assert.Equal(t, session.CompletedAt.Unix(), result.CompletedAt.Unix())
	session.StartedAt = time.Time{}
	result.StartedAt = time.Time{}
	session.CompletedAt = time.Time{}
	result.CompletedAt = time.Time{}
	session.cancelFunc = nil
	session.completed = nil
	assert.Equal(t, session, result)
}

type testStreamFactory struct {
}

func (sf *testStreamFactory) New(_, _ gopacket.Flow) tcpassembly.Stream {
	reader := tcpreader.NewReaderStream()
	go func() {
		buffer := bufio.NewReader(&reader)
		tcpreader.DiscardBytesToEOF(buffer)
	}()
	return &reader
}
