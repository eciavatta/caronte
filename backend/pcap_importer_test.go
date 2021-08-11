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

package main

import (
	"bufio"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/tcpassembly"
	"github.com/google/gopacket/tcpassembly/tcpreader"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestImportPcap(t *testing.T) {
	wrapper := NewTestStorageWrapper(t)
	pcapImporter := newTestPcapImporter(wrapper, "172.17.0.3")

	pcapImporter.releaseAssembler(pcapImporter.takeAssembler())

	fileName := copyToProcessing(t, "ping_pong_10000.pcap")
	sessionID, err := pcapImporter.ImportPcap(fileName, false)
	require.NoError(t, err)

	duplicatePcapFileName := copyToProcessing(t, "ping_pong_10000.pcap")
	duplicateSessionID, err := pcapImporter.ImportPcap(duplicatePcapFileName, false)
	require.Error(t, err)
	assert.Equal(t, EmptyRowID(), duplicateSessionID)
	assert.Error(t, os.Remove(ProcessingPcapsBasePath+duplicatePcapFileName))

	_, isPresent := pcapImporter.GetSession(EmptyRowID())
	assert.False(t, isPresent)

	session := waitSessionCompletion(t, pcapImporter, sessionID)
	assert.Equal(t, 15008, session.ProcessedPackets)
	assert.Equal(t, 0, session.InvalidPackets)
	assert.Equal(t, map[uint16]flowCount{9999: {10004, 5004}}, session.PacketsPerService)
	assert.Zero(t, session.ImportingError)
	assert.Equal(t, "369ef4b6abb6214b4ee2e0c81ecb93c49e275c26c85e30493b37727d408cf280", session.Hash)

	checkSessionEquals(t, wrapper, session)

	assert.Error(t, os.Remove(ProcessingPcapsBasePath+fileName))
	assert.NoError(t, os.Remove(PcapsBasePath+session.ID.Hex()+".pcap"))

	wrapper.Destroy(t)
}

func TestCancelImportSession(t *testing.T) {
	wrapper := NewTestStorageWrapper(t)
	pcapImporter := newTestPcapImporter(wrapper, "172.17.0.3")

	fileName := copyToProcessing(t, "ping_pong_10000.pcap")
	sessionID, err := pcapImporter.ImportPcap(fileName, false)
	require.NoError(t, err)

	assert.False(t, pcapImporter.CancelSession(EmptyRowID()))
	assert.True(t, pcapImporter.CancelSession(sessionID))

	session := waitSessionCompletion(t, pcapImporter, sessionID)
	assert.Zero(t, session.CompletedAt)
	assert.Equal(t, int64(1270696), session.Size)
	assert.Equal(t, 0, session.InvalidPackets)
	assert.NotZero(t, session.ImportingError)
	assert.Equal(t, "369ef4b6abb6214b4ee2e0c81ecb93c49e275c26c85e30493b37727d408cf280", session.Hash)

	checkSessionEquals(t, wrapper, session)

	assert.Error(t, os.Remove(ProcessingPcapsBasePath+fileName))
	assert.Error(t, os.Remove(PcapsBasePath+sessionID.Hex()+".pcap"))

	wrapper.Destroy(t)
}

func TestImportNoTcpPackets(t *testing.T) {
	wrapper := NewTestStorageWrapper(t)
	pcapImporter := newTestPcapImporter(wrapper, "172.17.0.4")

	fileName := copyToProcessing(t, "icmp.pcap")
	sessionID, err := pcapImporter.ImportPcap(fileName, false)
	require.NoError(t, err)

	session := waitSessionCompletion(t, pcapImporter, sessionID)
	assert.Equal(t, int64(228024), session.Size)
	assert.Equal(t, 2000, session.ProcessedPackets)
	assert.Equal(t, 2000, session.InvalidPackets)
	assert.Equal(t, map[uint16]flowCount{}, session.PacketsPerService)
	assert.Zero(t, session.ImportingError)
	assert.Equal(t, "392c71b41e6f1fc4333923ced430bd723d70b692c949c53e435d0db261386ee6", session.Hash)

	checkSessionEquals(t, wrapper, session)

	assert.Error(t, os.Remove(ProcessingPcapsBasePath+fileName))
	assert.NoError(t, os.Remove(PcapsBasePath+sessionID.Hex()+".pcap"))

	wrapper.Destroy(t)
}

func newTestPcapImporter(wrapper *TestStorageWrapper, serverAddress string) *PcapImporter {
	wrapper.AddCollection(ImportingSessions)

	streamPool := tcpassembly.NewStreamPool(&testStreamFactory{})

	notificationController := NewNotificationController(nil)
	go notificationController.Run()

	return &PcapImporter{
		storage:                wrapper.Storage,
		streamPool:             streamPool,
		assemblers:             make([]*tcpassembly.Assembler, 0, initialAssemblerPoolSize),
		sessions:               make(map[RowID]*ImportingSession),
		mAssemblers:            sync.Mutex{},
		mSessions:              sync.Mutex{},
		serverNet:              *ParseIPNet(serverAddress),
		notificationController: notificationController,
	}
}

func waitSessionCompletion(t *testing.T, pcapImporter *PcapImporter, sessionID RowID) ImportingSession {
	session, isPresent := pcapImporter.GetSession(sessionID)
	require.True(t, isPresent)
	count := 0
	for session.CompletedAt.IsZero() && count < 100 && session.ImportingError == "" {
		time.Sleep(100 * time.Millisecond)
		count++
		session, isPresent = pcapImporter.GetSession(sessionID)
	}
	assert.NotEqual(t, 100, count, "waitSessionCompletition timeout")

	assert.True(t, isPresent)
	assert.Equal(t, sessionID, session.ID)

	return session
}

func checkSessionEquals(t *testing.T, wrapper *TestStorageWrapper, session ImportingSession) {
	var result ImportingSession
	assert.NoError(t, wrapper.Storage.Find(ImportingSessions).
		Filter(OrderedDocument{{Key: "_id", Value: session.ID}}).
		Context(wrapper.Context).First(&result))
	assert.Equal(t, session.StartedAt.Unix(), result.StartedAt.Unix())
	assert.Equal(t, session.CompletedAt.Unix(), result.CompletedAt.Unix())
	session.StartedAt = time.Time{}
	result.StartedAt = time.Time{}
	session.CompletedAt = time.Time{}
	result.CompletedAt = time.Time{}
	session.cancelFunc = nil
	assert.Equal(t, session, result)
}

func copyToProcessing(t *testing.T, fileName string) string {
	newFile := fmt.Sprintf("test-%v-%s", time.Now().UnixNano(), fileName)
	require.NoError(t, CopyFile(ProcessingPcapsBasePath+newFile, "test_data/"+fileName))
	return newFile
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
