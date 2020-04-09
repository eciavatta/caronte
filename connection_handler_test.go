package main

import (
	"context"
	"encoding/binary"
	"github.com/flier/gohs/hyperscan"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net"
	"testing"
	"time"
)

func TestTakeReleaseScanners(t *testing.T) {
	wrapper := NewTestStorageWrapper(t)
	serverIP := layers.NewIPEndpoint(net.ParseIP(testDstIP))
	ruleManager := TestRuleManager{
		databaseUpdated: make(chan RulesDatabase),
	}

	database, err := hyperscan.NewStreamDatabase(hyperscan.NewPattern("/nope/", 0))
	require.NoError(t, err)

	factory := NewBiDirectionalStreamFactory(wrapper.Storage, serverIP, &ruleManager)
	version := wrapper.Storage.NewRowID()
	ruleManager.DatabaseUpdateChannel() <- RulesDatabase{database, 0, version}
	time.Sleep(10 * time.Millisecond)

	n := 100
	for i := 0; i < n; i++ {
		scanner := factory.takeScanner()
		assert.Equal(t, scanner.version, version)

		if i%5 == 0 {
			version = wrapper.Storage.NewRowID()
			ruleManager.DatabaseUpdateChannel() <- RulesDatabase{database, 0, version}
			time.Sleep(10 * time.Millisecond)
		}
		factory.releaseScanner(scanner)
	}
	assert.Len(t, factory.scanners, 1)

	scanners := make([]Scanner, n)
	for i := 0; i < n; i++ {
		scanners[i] = factory.takeScanner()
		assert.Equal(t, scanners[i].version, version)
	}

	version = wrapper.Storage.NewRowID()
	ruleManager.DatabaseUpdateChannel() <- RulesDatabase{database, 0, version}
	time.Sleep(10 * time.Millisecond)

	for i := 0; i < n; i++ {
		factory.releaseScanner(scanners[i])
	}
	assert.Len(t, factory.scanners, n)

	for i := 0; i < n; i++ {
		scanners[i] = factory.takeScanner()
		assert.Equal(t, scanners[i].version, version)
		factory.releaseScanner(scanners[i])
	}

	wrapper.Destroy(t)
}


func TestConnectionFactory(t *testing.T) {
	wrapper := NewTestStorageWrapper(t)
	wrapper.AddCollection(Connections)
	wrapper.AddCollection(ConnectionStreams)

	ruleManager := TestRuleManager{
		databaseUpdated: make(chan RulesDatabase),
	}

	serverIP := layers.NewIPEndpoint(net.ParseIP(testDstIP))
	clientIP := layers.NewIPEndpoint(net.ParseIP(testSrcIP))
	serverPort := layers.NewTCPPortEndpoint(dstPort)
	clientPort := layers.NewTCPPortEndpoint(srcPort)
	serverClientNetFlow, err := gopacket.FlowFromEndpoints(serverIP, clientIP)
	require.NoError(t, err)
	serverClientTransportFlow, err := gopacket.FlowFromEndpoints(serverPort, clientPort)
	require.NoError(t, err)
	clientServerNetFlow, err := gopacket.FlowFromEndpoints(clientIP, serverIP)
	require.NoError(t, err)
	clientServerTransportFlow, err := gopacket.FlowFromEndpoints(clientPort, serverPort)
	require.NoError(t, err)

	database, err := hyperscan.NewStreamDatabase(hyperscan.NewPattern("/nope/", 0))
	require.NoError(t, err)

	factory := NewBiDirectionalStreamFactory(wrapper.Storage, serverIP, &ruleManager)
	version := wrapper.Storage.NewRowID()
	ruleManager.DatabaseUpdateChannel() <- RulesDatabase{database, 0, version}
	time.Sleep(10 * time.Millisecond)

	serverStream := factory.New(serverClientNetFlow, serverClientTransportFlow)
	connectionFlow := StreamFlow{clientIP, serverIP, clientPort, serverPort}
	invertedConnectionFlow := StreamFlow{serverIP, clientIP, serverPort, clientPort}
	connection, isPresent := factory.connections[invertedConnectionFlow]
	require.True(t, isPresent)
	assert.Equal(t, connectionFlow, connection.(*connectionHandlerImpl).connectionFlow)

	serverStream.ReassemblyComplete()
	assert.Equal(t, invertedConnectionFlow, connection.(*connectionHandlerImpl).otherStream.streamFlow)

	clientStream := factory.New(clientServerNetFlow, clientServerTransportFlow)
	assert.Len(t, factory.connections, 0)
	clientStream.ReassemblyComplete()

	var result Connection
	err = wrapper.Storage.Find(Connections).Context(wrapper.Context).First(&result)
	require.NoError(t, err)

	assert.NotNil(t, result)
	assert.Equal(t, wrapper.Storage.NewCustomRowID(connectionFlow.Hash(), result.StartedAt), result.ID)
	assert.Equal(t, clientIP.String(), result.SourceIP)
	assert.Equal(t, serverIP.String(), result.DestinationIP)
	assert.Equal(t, binary.BigEndian.Uint16(clientPort.Raw()), result.SourcePort)
	assert.Equal(t, binary.BigEndian.Uint16(serverPort.Raw()), result.DestinationPort)

	wrapper.Destroy(t)
}


type TestRuleManager struct {
	databaseUpdated chan RulesDatabase
}

func (rm TestRuleManager) LoadRules() error {
	return nil
}

func (rm TestRuleManager) AddRule(_ context.Context, _ Rule) (string, error) {
	return "", nil
}

func (rm TestRuleManager) FillWithMatchedRules(_ *Connection, _ map[uint][]PatternSlice, _ map[uint][]PatternSlice) {
}

func (rm TestRuleManager) DatabaseUpdateChannel() chan RulesDatabase {
	return rm.databaseUpdated
}
