package main

import (
	"context"
	"encoding/binary"
	"github.com/flier/gohs/hyperscan"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/tcpassembly"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"math/rand"
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

	n := 1000
	for i := 0; i < n; i++ {
		scanner := factory.takeScanner()
		assert.Equal(t, scanner.version, version)

		if i%50 == 0 {
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
	for i := 0; i < n; i++ {
		factory.releaseScanner(scanners[i])
	}
	assert.Len(t, factory.scanners, n)

	version = wrapper.Storage.NewRowID()
	ruleManager.DatabaseUpdateChannel() <- RulesDatabase{database, 0, version}
	time.Sleep(10 * time.Millisecond)

	for i := 0; i < n; i++ {
		scanners[i] = factory.takeScanner()
		assert.Equal(t, scanners[i].version, version)
		factory.releaseScanner(scanners[i])
	}

	close(ruleManager.DatabaseUpdateChannel())
	wrapper.Destroy(t)
}

func TestConnectionFactory(t *testing.T) {
	wrapper := NewTestStorageWrapper(t)
	wrapper.AddCollection(Connections)
	wrapper.AddCollection(ConnectionStreams)

	ruleManager := TestRuleManager{
		databaseUpdated: make(chan RulesDatabase),
	}

	clientIP := layers.NewIPEndpoint(net.ParseIP(testSrcIP))
	serverIP := layers.NewIPEndpoint(net.ParseIP(testDstIP))
	serverPort := layers.NewTCPPortEndpoint(dstPort)
	clientServerNetFlow, err := gopacket.FlowFromEndpoints(clientIP, serverIP)
	require.NoError(t, err)
	serverClientNetFlow, err := gopacket.FlowFromEndpoints(serverIP, clientIP)
	require.NoError(t, err)

	database, err := hyperscan.NewStreamDatabase(hyperscan.NewPattern("/nope/", 0))
	require.NoError(t, err)

	factory := NewBiDirectionalStreamFactory(wrapper.Storage, serverIP, &ruleManager)
	version := wrapper.Storage.NewRowID()
	ruleManager.DatabaseUpdateChannel() <- RulesDatabase{database, 0, version}
	time.Sleep(10 * time.Millisecond)

	testInteraction := func(netFlow gopacket.Flow, transportFlow gopacket.Flow, otherSeenChan chan time.Time,
		completed chan bool) {

		time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)
		stream := factory.New(netFlow, transportFlow)
		seen := time.Now()
		stream.Reassembled([]tcpassembly.Reassembly{{[]byte{}, 0, true, true, seen}})
		stream.ReassemblyComplete()

		var startedAt, closedAt time.Time
		if netFlow == serverClientNetFlow {
			otherSeenChan <- seen
			return
		}

		otherSeen, ok := <-otherSeenChan
		require.True(t, ok)

		if seen.Before(otherSeen) {
			startedAt = seen
			closedAt = otherSeen
		} else {
			startedAt = otherSeen
			closedAt = seen
		}
		close(otherSeenChan)

		var result Connection
		connectionFlow := StreamFlow{netFlow.Src(), netFlow.Dst(), transportFlow.Src(), transportFlow.Dst()}
		connectionID := wrapper.Storage.NewCustomRowID(connectionFlow.Hash(), startedAt)
		op := wrapper.Storage.Find(Connections).Context(wrapper.Context)
		err := op.Filter(OrderedDocument{{"_id", connectionID}}).First(&result)
		require.NoError(t, err)

		assert.NotNil(t, result)
		assert.Equal(t, wrapper.Storage.NewCustomRowID(connectionFlow.Hash(), result.StartedAt), result.ID)
		assert.Equal(t, netFlow.Src().String(), result.SourceIP)
		assert.Equal(t, netFlow.Dst().String(), result.DestinationIP)
		assert.Equal(t, binary.BigEndian.Uint16(transportFlow.Src().Raw()), result.SourcePort)
		assert.Equal(t, binary.BigEndian.Uint16(transportFlow.Dst().Raw()), result.DestinationPort)
		assert.Equal(t, startedAt.Unix(), result.StartedAt.Unix())
		assert.Equal(t, closedAt.Unix(), result.ClosedAt.Unix())

		completed <- true
	}

	completed := make(chan bool)
	n := 1000

	for port := 40000; port < 40000+n; port++ {
		clientPort := layers.NewTCPPortEndpoint(layers.TCPPort(port))
		clientServerTransportFlow, err := gopacket.FlowFromEndpoints(clientPort, serverPort)
		require.NoError(t, err)
		serverClientTransportFlow, err := gopacket.FlowFromEndpoints(serverPort, clientPort)
		require.NoError(t, err)

		otherSeenChan := make(chan time.Time)
		go testInteraction(clientServerNetFlow, clientServerTransportFlow, otherSeenChan, completed)
		go testInteraction(serverClientNetFlow, serverClientTransportFlow, otherSeenChan, completed)
	}

	timeout := time.Tick(10 * time.Second)
	for i := 0; i < n; i++ {
		select {
		case <- completed:
			continue
		case <- timeout:
			t.Fatal("timeout")
		}
	}

	assert.Len(t, factory.connections, 0)

	close(ruleManager.DatabaseUpdateChannel())
	wrapper.Destroy(t)
}

type TestRuleManager struct {
	databaseUpdated chan RulesDatabase
}

func (rm TestRuleManager) LoadRules() error {
	return nil
}

func (rm TestRuleManager) AddRule(_ context.Context, _ Rule) (RowID, error) {
	return RowID{}, nil
}

func (rm TestRuleManager) GetRule(_ RowID) (Rule, bool) {
	return Rule{}, false
}

func (rm TestRuleManager) UpdateRule(_ context.Context, _ Rule) bool {
	return false
}

func (rm TestRuleManager) GetRules() []Rule {
	return nil
}

func (rm TestRuleManager) FillWithMatchedRules(_ *Connection, _ map[uint][]PatternSlice, _ map[uint][]PatternSlice) {
}

func (rm TestRuleManager) DatabaseUpdateChannel() chan RulesDatabase {
	return rm.databaseUpdated
}
