package main

import (
	"bufio"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/tcpassembly"
	"github.com/google/gopacket/tcpassembly/tcpreader"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/atomic"
	"net"
	"sync"
	"testing"
)

func TestImportPcap(t *testing.T) {
	wrapper := NewTestStorageWrapper(t)
	wrapper.AddCollection(ImportingSessions)

	serverEndpoint := layers.NewIPEndpoint(net.ParseIP("172.17.0.3"))
	streamPool := tcpassembly.NewStreamPool(&testStreamFactory{})

	pcapImporter := PcapImporter{
		storage:     wrapper.Storage,
		streamPool:  streamPool,
		assemblers:  make([]*tcpassembly.Assembler, 0, initialAssemblerPoolSize),
		sessions:    make(map[string]ImportingSession),
		mAssemblers: sync.Mutex{},
		mSessions:   sync.Mutex{},
		serverIP:    serverEndpoint,
	}

	sessionID, err := pcapImporter.ImportPcap("test_data/ping_pong_10000.pcap")
	require.NoError(t, err)
	assert.NotZero(t, sessionID)

	duplicateSessionID, err := pcapImporter.ImportPcap("test_data/ping_pong_10000.pcap")
	require.Error(t, err)
	assert.Equal(t, sessionID, duplicateSessionID)

	_, isPresent := pcapImporter.GetSession("invalid")
	assert.False(t, isPresent)

	session, isPresent := pcapImporter.GetSession(sessionID)
	require.True(t, isPresent)
	err, _ = <- session.completed

	session, isPresent = pcapImporter.GetSession(sessionID)
	require.True(t, isPresent)
	assert.NoError(t, err)
	assert.Equal(t, sessionID, session.ID)
	assert.Equal(t, 15008, session.ProcessedPackets)
	assert.Equal(t, 0, session.InvalidPackets)
	assert.Equal(t, map[uint16]flowCount{9999: {10004, 5004}}, session.PacketsPerService)
	assert.NoError(t, session.ImportingError)

	wrapper.Destroy(t)
}

type testStreamFactory struct{
	counter atomic.Int32
}

func (sf *testStreamFactory) New(_, _ gopacket.Flow) tcpassembly.Stream {
	sf.counter.Inc()
	reader := tcpreader.NewReaderStream()
	go func() {
		buffer := bufio.NewReader(&reader)
		tcpreader.DiscardBytesToEOF(buffer)
	}()
	return &reader
}
