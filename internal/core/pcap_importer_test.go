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
	"bufio"
	"fmt"
	"net/http"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/gopacket"
	"github.com/google/gopacket/tcpassembly"
	"github.com/google/gopacket/tcpassembly/tcpreader"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestImportPcap(t *testing.T) {
	wrapper := NewTestStorageWrapper(t)
	pcapImporter, statsChan := newTestPcapImporter(wrapper, "172.17.0.3")

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

	assert.Equal(t, "pcap.completed", (<-statsChan)["event"])
	assert.Equal(t, gin.H{"event": "packets.statistics", "message": PacketsStatistics{
		ProcessedPackets: 15008,
		InvalidPackets:   0,
		PacketsPerMinute: 0,
	}}, <-statsChan)

	assert.Error(t, os.Remove(ProcessingPcapsBasePath+fileName))
	assert.NoError(t, os.Remove(PcapsBasePath+session.ID.Hex()+".pcap"))

	wrapper.Destroy(t)
}

func TestCancelImportSession(t *testing.T) {
	wrapper := NewTestStorageWrapper(t)
	pcapImporter, statsChan := newTestPcapImporter(wrapper, "172.17.0.3")

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

	assert.Equal(t, "pcap.canceled", (<-statsChan)["event"])
	time.Sleep(time.Second)

	assert.Error(t, os.Remove(ProcessingPcapsBasePath+fileName))
	assert.Error(t, os.Remove(PcapsBasePath+sessionID.Hex()+".pcap"))

	wrapper.Destroy(t)
}

func TestImportNoTcpPackets(t *testing.T) {
	wrapper := NewTestStorageWrapper(t)
	pcapImporter, statsChan := newTestPcapImporter(wrapper, "172.17.0.4")

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

	time.Sleep(time.Second) // wait to write session on database
	checkSessionEquals(t, wrapper, session)

	assert.Equal(t, "pcap.completed", (<-statsChan)["event"])
	assert.Equal(t, gin.H{"event": "packets.statistics", "message": PacketsStatistics{
		ProcessedPackets: 2000,
		InvalidPackets:   2000,
		PacketsPerMinute: 0,
	}}, <-statsChan)

	assert.Error(t, os.Remove(ProcessingPcapsBasePath+fileName))
	assert.NoError(t, os.Remove(PcapsBasePath+sessionID.Hex()+".pcap"))

	wrapper.Destroy(t)
}

func TestListInterfaces(t *testing.T) {
	pcapImporter, _ := newTestPcapImporter(nil, "127.0.0.1")

	interfaces, err := pcapImporter.ListInterfaces()
	require.NoError(t, err)
	assert.NotEmpty(t, interfaces)
	assert.Contains(t, interfaces, "lo", "loopback interface must always be present")
}

func TestListRemoteInterfaces(t *testing.T) {
	pcapImporter, _ := newTestPcapImporter(nil, "127.0.0.1")

	interfaces, err := pcapImporter.ListRemoteInterfaces(validSSHConfig())
	require.NoError(t, err)
	assert.Contains(t, interfaces, "lo")
	assert.Contains(t, interfaces, "eth0")
}

func TestRemoteSSHConnections(t *testing.T) {
	pcapImporter, _ := newTestPcapImporter(nil, "127.0.0.1")

	_, err := pcapImporter.ListRemoteInterfaces(validSSHConfig())
	require.NoError(t, err)

	_, err = pcapImporter.ListRemoteInterfaces(SSHConfig{
		Host:     "invalid",
		Password: "invalid",
	})
	require.Error(t, err)

	_, err = pcapImporter.ListRemoteInterfaces(SSHConfig{
		Host:     "1.1.1.1",
		Password: "invalid",
	})
	require.Error(t, err)

	_, err = pcapImporter.ListRemoteInterfaces(SSHConfig{
		Host:     testEnvironmentHost(),
		Password: "invalid",
	})
	require.Error(t, err)

	_, err = pcapImporter.ListRemoteInterfaces(SSHConfig{
		Host: testEnvironmentHost(),
		Port: testEnvironmentSshPort(),
	})
	require.Error(t, err)

	_, err = pcapImporter.ListRemoteInterfaces(SSHConfig{
		Host:     testEnvironmentHost(),
		Port:     testEnvironmentSshPort(),
		User:     "invalid",
		Password: "test",
	})
	require.Error(t, err)

	_, err = pcapImporter.ListRemoteInterfaces(SSHConfig{
		Host:     testEnvironmentHost(),
		Port:     testEnvironmentSshPort(),
		Password: "invalid",
	})
	require.Error(t, err)

	_, err = pcapImporter.ListRemoteInterfaces(SSHConfig{
		Host:       testEnvironmentHost(),
		Port:       testEnvironmentSshPort(),
		PrivateKey: "invalid",
	})
	require.Error(t, err)

	_, err = pcapImporter.ListRemoteInterfaces(SSHConfig{
		Host:       testEnvironmentHost(),
		Port:       testEnvironmentSshPort(),
		PrivateKey: "-----BEGIN OPENSSH PRIVATE KEY-----\nb3BlbnNzaC1rZXktdjEAAAAACmFlczI1Ni1jdHIAAAAGYmNyeXB0AAAAGAAAABAmwQugAw\nGA4R4hFZj7qrsIAAAAZAAAAAEAAAAzAAAAC3NzaC1lZDI1NTE5AAAAIHg+iDsYjtx9Y7iD\nCzHX0xoHWaTDA6fVs2CVXnDI7DWFAAAAsBwIgJfamwWHbdYIyMjAfogJO6Nt1BbxrAlFyW\nHObm8k7OfIc8iAdlIeUDtV9RvWtTVF6URIXZfxGxzpzXnVIBBwZkqR9zI8dB6RP7rR0t3D\nD3P1yFtXz2ei1ssa1ueoGV/0pojClroc+WKZJZD4qCGYDJ/vagy2ZSoOGJxgRoFGFRtuUx\n/FWilPP8urJQcnu4eDYqaZAfp+YS8QnlbBfnYfFSCSPiGWarxmYbdw0kwr\n-----END OPENSSH PRIVATE KEY-----\n",
	})
	require.Error(t, err, "key require passphrase")

	_, err = pcapImporter.ListRemoteInterfaces(SSHConfig{
		Host:       testEnvironmentHost(),
		Port:       testEnvironmentSshPort(),
		PrivateKey: "invalid",
	})
	require.Error(t, err)

	_, err = pcapImporter.ListRemoteInterfaces(SSHConfig{
		Host:       testEnvironmentHost(),
		Port:       testEnvironmentSshPort(),
		PrivateKey: "-----BEGIN OPENSSH PRIVATE KEY-----\nb3BlbnNzaC1rZXktdjEAAAAABG5vbmUAAAAEbm9uZQAAAAAAAAABAAAAMwAAAAtzc2gtZW\nQyNTUxOQAAACDpvMt45cHBpPQ6+MbQSPbqX/M2PvpeIDjXU3TxYMGUiQAAAJgOkOvQDpDr\n0AAAAAtzc2gtZWQyNTUxOQAAACDpvMt45cHBpPQ6+MbQSPbqX/M2PvpeIDjXU3TxYMGUiQ\nAAAEBBk7B4xNF0kG6w+sw7kuTsQyvc3wrey+q4SjcYZzNpb+m8y3jlwcGk9Dr4xtBI9upf\n8zY++l4gONdTdPFgwZSJAAAAFWNhcm9udGVAZWNpYXZhdHRhLmRldg==\n-----END OPENSSH PRIVATE KEY-----\n",
	})
	require.NoError(t, err)

	_, err = pcapImporter.ListRemoteInterfaces(SSHConfig{
		Host:       testEnvironmentHost(),
		Port:       testEnvironmentSshPort(),
		PrivateKey: "-----BEGIN OPENSSH PRIVATE KEY-----\nb3BlbnNzaC1rZXktdjEAAAAACmFlczI1Ni1jdHIAAAAGYmNyeXB0AAAAGAAAABAmwQugAw\nGA4R4hFZj7qrsIAAAAZAAAAAEAAAAzAAAAC3NzaC1lZDI1NTE5AAAAIHg+iDsYjtx9Y7iD\nCzHX0xoHWaTDA6fVs2CVXnDI7DWFAAAAsBwIgJfamwWHbdYIyMjAfogJO6Nt1BbxrAlFyW\nHObm8k7OfIc8iAdlIeUDtV9RvWtTVF6URIXZfxGxzpzXnVIBBwZkqR9zI8dB6RP7rR0t3D\nD3P1yFtXz2ei1ssa1ueoGV/0pojClroc+WKZJZD4qCGYDJ/vagy2ZSoOGJxgRoFGFRtuUx\n/FWilPP8urJQcnu4eDYqaZAfp+YS8QnlbBfnYfFSCSPiGWarxmYbdw0kwr\n-----END OPENSSH PRIVATE KEY-----\n",
		Passphrase: "test",
	})
	require.NoError(t, err)
}

func TestRemoteCapture(t *testing.T) {
	wrapper := NewTestStorageWrapper(t)
	pcapImporter, _ := newTestPcapImporter(wrapper, "172.0.0.0/8")

	validCaptureOptions := CaptureOptions{
		Interface:        "eth0",
		IncludedServices: []uint16{testEnvironmentHttpPort()},
	}

	err := pcapImporter.StartRemoteCapture(invalidSSHConfig(), validCaptureOptions)
	require.Error(t, err)

	err = pcapImporter.StartRemoteCapture(validSSHConfig(), CaptureOptions{
		Interface: "invalid",
	})
	require.Error(t, err)

	for i := 0; i < 3; i++ {
		require.NoError(t, pcapImporter.StartRemoteCapture(validSSHConfig(), validCaptureOptions))
		// one session per time
		require.Error(t, pcapImporter.StartRemoteCapture(validSSHConfig(), validCaptureOptions))

		time.Sleep(1 * time.Second)
		resp, err := http.Get(fmt.Sprintf("http://%s:%v/numbers", testEnvironmentHost(), testEnvironmentHttpPort()))
		assert.NoError(t, err)
		assert.Equal(t, resp.StatusCode, http.StatusOK)

		time.Sleep(2 * time.Second)
		require.NoError(t, pcapImporter.StopCapture())
		require.Error(t, pcapImporter.StopCapture())

		time.Sleep(2 * time.Second)
		sessions := pcapImporter.GetSessions()
		assert.Len(t, sessions, i+1)
		assert.Zero(t, sessions[i].ImportingError)
		assert.Less(t, 5, sessions[i].ProcessedPackets)
		assert.Less(t, 5, sessions[i].InvalidPackets) // todo: why invalid?

		assert.NoError(t, os.Remove(PcapsBasePath+sessions[i].ID.Hex()+".pcap"))
	}

	wrapper.Destroy(t)
}

func TestPcapRotation(t *testing.T) {
	wrapper := NewTestStorageWrapper(t)
	pcapImporter, _ := newTestPcapImporter(wrapper, "172.0.0.0/8")

	pcapImporter.SetSessionRotationInterval(3 * time.Second)

	require.NoError(t, pcapImporter.StartRemoteCapture(validSSHConfig(), CaptureOptions{
		Interface:        "eth0",
		IncludedServices: []uint16{testEnvironmentHttpPort()},
	}))

	time.Sleep(time.Second)

	for i := 0; i < 5; i++ {
		resp, err := http.Get(fmt.Sprintf("http://%s:%v/numbers", testEnvironmentHost(), testEnvironmentHttpPort()))
		assert.NoError(t, err)
		assert.Equal(t, resp.StatusCode, http.StatusOK)

		if i == 4 {
			time.Sleep(1 * time.Second)
			require.NoError(t, pcapImporter.StopCapture())
		}

		time.Sleep(3 * time.Second)
		sessions := pcapImporter.GetSessions()
		assert.Len(t, sessions, i+1)
		assert.Zero(t, sessions[i].ImportingError)
		assert.Less(t, 5, sessions[i].ProcessedPackets)
		assert.Less(t, 5, sessions[i].InvalidPackets) // todo: why invalid?

		assert.NoError(t, os.Remove(PcapsBasePath+sessions[i].ID.Hex()+".pcap"))
	}

	wrapper.Destroy(t)
}

func newTestPcapImporter(wrapper *TestStorageWrapper, serverAddress string) (*PcapImporter, chan gin.H) {
	var mongoStorage *MongoStorage
	if wrapper != nil {
		mongoStorage = wrapper.Storage
		wrapper.AddCollection(ImportingSessions)
	}

	streamPool := tcpassembly.NewStreamPool(&testStreamFactory{})

	notificationController := NewTestNotificationController()

	pcapImporter := &PcapImporter{
		storage:                 mongoStorage,
		streamPool:              streamPool,
		assemblers:              make([]*tcpassembly.Assembler, 0, initialAssemblerPoolSize),
		sessions:                make(map[RowID]*ImportingSession),
		mAssemblers:             sync.Mutex{},
		mSessions:               sync.Mutex{},
		serverNet:               *ParseIPNet(serverAddress),
		notificationController:  notificationController,
		mLiveCapture:            sync.Mutex{},
		sessionRotationInterval: initialSessionRotationInterval,
		packetsStatusChannel:    make(chan bool),
	}
	go pcapImporter.notificationService()

	return pcapImporter, notificationController.notificationChannel
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

func validSSHConfig() SSHConfig {
	return SSHConfig{
		Host:     testEnvironmentHost(),
		Port:     testEnvironmentSshPort(),
		User:     "root",
		Password: "test",
	}
}

func invalidSSHConfig() SSHConfig {
	return SSHConfig{
		Host:     testEnvironmentHost(),
		Port:     testEnvironmentSshPort(),
		User:     "root",
		Password: "wrong",
	}
}
