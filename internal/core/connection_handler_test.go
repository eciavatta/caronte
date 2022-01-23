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
	"context"
	"encoding/binary"
	"math/rand"
	"net"
	"testing"
	"time"

	"github.com/eciavatta/caronte/pkg/tcpassembly"
	"github.com/flier/gohs/hyperscan"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTakeReleaseScanners(t *testing.T) {
	wrapper := NewTestStorageWrapper(t)
	serverNet := ParseIPNet(testDstIP)
	ruleManager := TestRulesManager{
		databaseUpdated: make(chan RulesDatabase),
	}

	database, err := hyperscan.NewStreamDatabase(hyperscan.NewPattern("/nope/", 0))
	require.NoError(t, err)

	factory := NewBiDirectionalStreamFactory(wrapper.Storage, serverNet, &ruleManager, nil)
	version := NewRowID()
	ruleManager.DatabaseUpdateChannel() <- RulesDatabase{database, 0, version}
	time.Sleep(10 * time.Millisecond)

	n := 1000
	for i := 0; i < n; i++ {
		scanner := factory.takeScanner()
		assert.Equal(t, scanner.version, version)

		if i%50 == 0 {
			version = NewRowID()
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

	version = NewRowID()
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

	ruleManager := TestRulesManager{
		databaseUpdated: make(chan RulesDatabase),
	}
	notificationController := NewTestNotificationController()

	clientIP := layers.NewIPEndpoint(net.ParseIP(testSrcIP))
	serverIP := layers.NewIPEndpoint(net.ParseIP(testDstIP))
	serverPort := layers.NewTCPPortEndpoint(dstPort)
	clientServerNetFlow, err := gopacket.FlowFromEndpoints(clientIP, serverIP)
	require.NoError(t, err)
	serverClientNetFlow, err := gopacket.FlowFromEndpoints(serverIP, clientIP)
	require.NoError(t, err)

	database, err := hyperscan.NewStreamDatabase(hyperscan.NewPattern("/nope/", 0))
	require.NoError(t, err)

	factory := NewBiDirectionalStreamFactory(wrapper.Storage, ParseIPNet(testDstIP), &ruleManager, notificationController)
	version := NewRowID()
	ruleManager.DatabaseUpdateChannel() <- RulesDatabase{database, 0, version}
	time.Sleep(10 * time.Millisecond)

	testInteraction := func(netFlow gopacket.Flow, transportFlow gopacket.Flow, otherSeenChan chan time.Time,
		isServer bool, completed chan bool) {

		time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)
		stream := factory.New(netFlow, transportFlow, isServer)
		seen := time.Now()
		stream.Reassembled([]tcpassembly.Reassembly{{
			Bytes: []byte{},
			Skip:  0,
			Start: true,
			End:   true,
			Seen:  seen,
		}})
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
		connectionID := CustomRowID(connectionFlow.Hash(), startedAt)
		op := wrapper.Storage.Find(Connections).Context(wrapper.Context)
		err := op.Filter(OrderedDocument{{Key: "_id", Value: connectionID}}).First(&result)
		require.NoError(t, err)

		assert.NotNil(t, result)
		assert.Equal(t, CustomRowID(connectionFlow.Hash(), result.StartedAt), result.ID)
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
		go testInteraction(clientServerNetFlow, clientServerTransportFlow, otherSeenChan, false, completed)
		go testInteraction(serverClientNetFlow, serverClientTransportFlow, otherSeenChan, true, completed)
	}

	timeout := time.Tick(30 * time.Second)
	for i := 0; i < n; i++ {
		select {
		case <-completed:
			continue
		case message := <-notificationController.notificationChannel:
			assert.Equal(t, "connections.statistics", message["event"])
		case <-timeout:
			t.Fatal("timeout")
		}
	}

	message := <-notificationController.notificationChannel
	assert.Equal(t, "connections.statistics", message["event"])
	assert.Equal(t, ConnectionsStatistics{
		PendingConnections:   0,
		CompletedConnections: uint(n),
		ConnectionsPerMinute: 0, // updated after 1 minute, but test end in few seconds
	}, message["message"])

	assert.Len(t, factory.connections, 0)

	close(ruleManager.DatabaseUpdateChannel())
	wrapper.Destroy(t)
}

type TestRulesManager struct {
	databaseUpdated chan RulesDatabase
}

func (rm TestRulesManager) LoadRules() error {
	return nil
}

func (rm TestRulesManager) AddRule(_ context.Context, _ Rule) (RowID, error) {
	return RowID{}, nil
}

func (rm TestRulesManager) GetRule(_ RowID) (Rule, bool) {
	return Rule{}, false
}

func (rm TestRulesManager) UpdateRule(_ context.Context, _ RowID, _ Rule) (bool, error) {
	return false, nil
}

func (rm TestRulesManager) GetRules() []Rule {
	return nil
}

func (rm TestRulesManager) SetFlag(_ context.Context, _ string) error {
	return nil
}

func (rm TestRulesManager) FillWithMatchedRules(_ *Connection, _ map[uint][]PatternSlice, _ map[uint][]PatternSlice) {
}

func (rm TestRulesManager) DatabaseUpdateChannel() chan RulesDatabase {
	return rm.databaseUpdated
}
