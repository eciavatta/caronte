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
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"github.com/google/gopacket/pcapgo"
	"github.com/google/gopacket/tcpassembly"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
)

const PcapsBasePath = "pcaps/"
const ProcessingPcapsBasePath = PcapsBasePath + "processing/"
const initialAssemblerPoolSize = 16
const initialSessionRotationInterval = 2 * time.Minute
const snapshotLen = 1024
const remoteCapturePipeName = "remote-pipe.pcap"

type PcapImporter struct {
	storage                 Storage
	streamPool              *tcpassembly.StreamPool
	assemblers              []*tcpassembly.Assembler
	sessions                map[RowID]*ImportingSession
	mAssemblers             sync.Mutex
	mSessions               sync.Mutex
	serverNet               net.IPNet
	notificationController  *NotificationController
	liveCaptureHandle       *pcap.Handle
	mLiveCapture            sync.Mutex
	currentLiveSession      *ImportingSession
	sessionRotationInterval time.Duration
}

type ImportingSession struct {
	ID                RowID                `json:"id" bson:"_id"`
	Hash              string               `json:"hash" bson:"hash"`
	StartedAt         time.Time            `json:"started_at" bson:"started_at"`
	Size              int64                `json:"size" bson:"size"`
	CompletedAt       time.Time            `json:"completed_at" bson:"completed_at,omitempty"`
	ProcessedPackets  int                  `json:"processed_packets" bson:"processed_packets"`
	InvalidPackets    int                  `json:"invalid_packets" bson:"invalid_packets"`
	PacketsPerService map[uint16]flowCount `json:"packets_per_service" bson:"packets_per_service"`
	ImportingError    string               `json:"importing_error" bson:"importing_error,omitempty"`
	cancelFunc        context.CancelFunc
}

type CaptureOptions struct {
	Interface        string   `json:"interface" binding:"required"`
	IncludedServices []uint16 `json:"included_services"`
	ExcludedServices []uint16 `json:"excluded_services"`
}

type SSHConfig struct {
	Host            string `json:"host" binding:"required"`
	Port            uint16 `json:"port"`
	User            string `json:"user"`
	Password        string `json:"password"`
	PrivateKey      string `json:"private_key"`
	Passphrase      string `json:"passphrase"`
	ServerPublicKey string `json:"server_public_key"`
}

type flowCount [2]int

func NewPcapImporter(storage Storage, serverNet net.IPNet, rulesManager RulesManager,
	notificationController *NotificationController) *PcapImporter {
	streamPool := tcpassembly.NewStreamPool(NewBiDirectionalStreamFactory(storage, serverNet, rulesManager))

	var result []ImportingSession
	if err := storage.Find(ImportingSessions).All(&result); err != nil {
		log.WithError(err).Panic("failed to retrieve importing sessions")
	}
	sessions := make(map[RowID]*ImportingSession)
	for _, session := range result {
		sessions[session.ID] = &session
	}

	return &PcapImporter{
		storage:                 storage,
		streamPool:              streamPool,
		assemblers:              make([]*tcpassembly.Assembler, 0, initialAssemblerPoolSize),
		sessions:                sessions,
		mAssemblers:             sync.Mutex{},
		mSessions:               sync.Mutex{},
		serverNet:               serverNet,
		notificationController:  notificationController,
		mLiveCapture:            sync.Mutex{},
		sessionRotationInterval: initialSessionRotationInterval,
	}
}

// Import a pcap file to the database. The pcap file must be present at the fileName path. If the pcap is already
// going to be imported or if it has been already imported in the past the function returns an error. Otherwise it
// create a new session and starts to import the pcap, and returns immediately the session name (that is the sha256
// of the pcap).
func (pi *PcapImporter) ImportPcap(fileName string, flushAll bool) (RowID, error) {
	switch filepath.Ext(fileName) {
	case ".pcap":
	case ".pcapng":
	default:
		deleteProcessingFile(fileName)
		return EmptyRowID(), errors.New("invalid file extension")
	}

	hash, err := Sha256Sum(ProcessingPcapsBasePath + fileName)
	if err != nil {
		deleteProcessingFile(fileName)
		log.WithError(err).Panic("failed to calculate pcap sha256")
	}

	pi.mSessions.Lock()
	isPresent := false
	for _, session := range pi.sessions {
		if session.Hash == hash {
			isPresent = true
			break
		}
	}
	if isPresent {
		pi.mSessions.Unlock()
		deleteProcessingFile(fileName)
		return EmptyRowID(), errors.New("pcap already processed")
	}

	handle, err := pcap.OpenOffline(ProcessingPcapsBasePath + fileName)
	if err != nil {
		pi.mSessions.Unlock()
		deleteProcessingFile(fileName)
		log.WithError(err).Panic("failed to process pcap")
	}

	session, ctx := pi.newSession(false)
	session.Hash = hash
	session.Size = FileSize(ProcessingPcapsBasePath + fileName)

	pi.sessions[session.ID] = session
	pi.mSessions.Unlock()

	go pi.handle(handle, session, flushAll, ctx, func(canceled bool) {
		handle.Close()
		if canceled {
			deleteProcessingFile(fileName)
		} else {
			moveProcessingFile(session, fileName)
		}
	})

	return session.ID, nil
}

func (pi *PcapImporter) StartCapturing(captureOptions CaptureOptions) error {
	pi.mLiveCapture.Lock()
	defer pi.mLiveCapture.Unlock()

	if pi.liveCaptureHandle != nil {
		return errors.New("live capture is already in progress")
	}

	handle, err := pcap.OpenLive(captureOptions.Interface, 1600, true, pcap.BlockForever)
	if err != nil {
		return err
	}

	if err := handle.SetBPFFilter(generateBffFilters(captureOptions)); err != nil {
		return err
	}

	pi.liveCaptureHandle = handle
	go pi.handle(handle, nil, true, nil, func(_ bool) {
		handle.Close()
	})

	return nil
}

func (pi *PcapImporter) StopCapturing() error {
	pi.mLiveCapture.Lock()
	defer pi.mLiveCapture.Unlock()

	if pi.liveCaptureHandle == nil {
		return errors.New("live capture is already stopped")
	}

	pi.currentLiveSession.cancelFunc()
	pi.liveCaptureHandle = nil

	return nil
}

func (pi *PcapImporter) ListInterfaces() ([]string, error) {
	ifs, err := pcap.FindAllDevs()
	if err != nil {
		return nil, err
	}

	interfaces := make([]string, len(ifs))
	for i := range ifs {
		interfaces[i] = ifs[i].Name
	}

	return interfaces, nil
}

func (pi *PcapImporter) StartRemoteCapturing(sshConfig SSHConfig, captureOptions CaptureOptions) error {
	client, err := createSSHClient(&sshConfig)
	if err != nil {
		return err
	}

	// check if tcpdump is present
	session, err := client.NewSession()
	if err != nil {
		return err
	}
	if err = session.Run("tcpdump --version"); err != nil {
		return err
	}
	if err = session.Close(); err != nil && err != io.EOF {
		log.WithError(err).Panic("failed to close ssh session")
	}

	if FileExists(remoteCapturePipeName) {
		deleteProcessingFile(remoteCapturePipeName)
	}

	pipeName := ProcessingPcapsBasePath + remoteCapturePipeName
	if err = syscall.Mkfifo(pipeName, 0666); err != nil {
		log.WithError(err).Panic("failed to create named pipe")
	}

	namedPipe, err := os.OpenFile(pipeName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0777)
	if err != nil {
		log.WithError(err).Panic("failed to open named pipe")
	}

	session, err = client.NewSession()
	if err != nil {
		return err
	}
	session.Stdout = bufio.NewWriter(namedPipe)
	captureOptions.ExcludedServices = append(captureOptions.ExcludedServices, sshConfig.Port)

	go (func() {
		if err = session.Run(fmt.Sprintf("tcpdump -s 0 -U -n -w - -i %s %s",
			captureOptions.Interface, generateBffFilters(captureOptions))); err != nil {
			log.WithError(err).Panic("failed to start tcpdump on remote host")
		}
		if err = session.Close(); err != nil && err != io.EOF {
			log.WithError(err).Panic("failed to close ssh session")
		}
	})()

	handle, err := pcap.OpenOfflineFile(namedPipe)
	if err != nil {
		return err
	}

	go pi.handle(handle, nil, true, nil, func(canceled bool) {
		handle.Close()
		deleteProcessingFile(remoteCapturePipeName)
	})

	return nil
}

func (pi *PcapImporter) ListRemoteInterfaces(sshConfig SSHConfig) ([]string, error) {
	client, err := createSSHClient(&sshConfig)
	if err != nil {
		return nil, err
	}

	session, err := client.NewSession()
	if err != nil {
		return nil, err
	}

	var interfaces []string
	if output, err := session.CombinedOutput("/bin/ls /sys/class/net/"); err != nil {
		return nil, err
	} else {
		interfaces = strings.Split(string(output), "\n")
	}
	if err = session.Close(); err != nil && err != io.EOF {
		log.WithError(err).Panic("failed to close ssh session")
	}

	return interfaces[:len(interfaces)-1], nil
}

func (pi *PcapImporter) GetSessions() []ImportingSession {
	pi.mSessions.Lock()
	sessions := make([]ImportingSession, 0, len(pi.sessions))
	for _, session := range pi.sessions {
		sessions = append(sessions, *session)
	}
	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].StartedAt.Before(sessions[j].StartedAt)
	})
	pi.mSessions.Unlock()
	return sessions
}

func (pi *PcapImporter) GetSession(sessionID RowID) (ImportingSession, bool) {
	pi.mSessions.Lock()
	defer pi.mSessions.Unlock()
	if session, isPresent := pi.sessions[sessionID]; isPresent {
		return *session, true
	} else {
		return ImportingSession{}, false
	}
}

func (pi *PcapImporter) CancelSession(sessionID RowID) bool {
	pi.mSessions.Lock()
	session, isPresent := pi.sessions[sessionID]
	if isPresent {
		session.cancelFunc()
	}
	pi.mSessions.Unlock()
	return isPresent
}

func (pi *PcapImporter) FlushConnections(olderThen time.Time, closeAll bool) (flushed, closed int) {
	assembler := pi.takeAssembler()
	flushed, closed = assembler.FlushWithOptions(tcpassembly.FlushOptions{
		T:        olderThen,
		CloseAll: closeAll,
	})
	pi.releaseAssembler(assembler)
	return
}

func (pi *PcapImporter) SetSessionRotationInterval(interval time.Duration) {
	pi.mLiveCapture.Lock()
	pi.sessionRotationInterval = interval
	pi.mLiveCapture.Unlock()
}

func (pi *PcapImporter) handle(handle *pcap.Handle, initialSession *ImportingSession,
	flushAll bool, ctx context.Context, onComplete func(canceled bool)) {
	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
	packetSource.NoCopy = true
	assembler := pi.takeAssembler()
	packets := packetSource.Packets()

	var currentFile *os.File
	var currentWriter *pcapgo.Writer

	sessionRotationInterval := time.Tick(pi.sessionRotationInterval)
	isOnline := false
	session := initialSession
	if initialSession == nil {
		isOnline = true
		session, ctx = pi.newSession(true)
		currentFile, currentWriter = createNewPcap(handle)
	}

	offlineLock := func() {
		if !isOnline {
			pi.mSessions.Lock()
		}
	}
	offlineUnlock := func() {
		if !isOnline {
			pi.mSessions.Unlock()
		}
	}

	rotateSession := func(isEnd bool) {
		if isOnline {
			savePcap(session, currentFile)

			pi.mSessions.Lock()
			pi.sessions[session.ID] = session
			pi.mSessions.Unlock()

			log.WithField("id", session.ID).WithField("hash", session.Hash).Debug("session rotated")

			if !isEnd {
				pi.notificationController.Notify("pcap.rotation", session)
				pi.saveSession(session, "")
				session, ctx = pi.newSession(true)
				currentFile, currentWriter = createNewPcap(handle)
			}
		}
	}

	flushAllIfNecessary := func() {
		if flushAll {
			connectionsClosed := assembler.FlushAll()
			log.Debugf("connections closed after flush: %v", connectionsClosed)
		}
	}

	for {
		select {
		case packet := <-packets:
			if packet == nil { // completed
				flushAllIfNecessary()
				pi.releaseAssembler(assembler)
				pi.saveSession(session, "")
				pi.notificationController.Notify("pcap.completed", session)
				onComplete(false)

				return
			}

			offlineLock()
			session.ProcessedPackets++

			if packet.NetworkLayer() == nil || packet.TransportLayer() == nil ||
				packet.TransportLayer().LayerType() != layers.LayerTypeTCP { // invalid packet
				session.InvalidPackets++
				offlineUnlock()

				continue
			}

			tcp := packet.TransportLayer().(*layers.TCP)
			var servicePort uint16
			var index int

			isDstServer := pi.serverNet.Contains(packet.NetworkLayer().NetworkFlow().Dst().Raw())
			isSrcServer := pi.serverNet.Contains(packet.NetworkLayer().NetworkFlow().Src().Raw())
			if isDstServer && !isSrcServer {
				servicePort = uint16(tcp.DstPort)
				index = 0
			} else if isSrcServer && !isDstServer {
				servicePort = uint16(tcp.SrcPort)
				index = 1
			} else {
				session.InvalidPackets++
				offlineUnlock()

				continue
			}
			fCount, isPresent := session.PacketsPerService[servicePort]
			if !isPresent {
				fCount = flowCount{0, 0}
			}
			fCount[index]++
			session.PacketsPerService[servicePort] = fCount
			offlineUnlock()

			if isOnline {
				currentWriter.WritePacket(packet.Metadata().CaptureInfo, packet.Data())
			}

			assembler.AssembleWithTimestamp(packet.NetworkLayer().NetworkFlow(), tcp, packet.Metadata().Timestamp)
		case <-sessionRotationInterval:
			rotateSession(false)
		case <-ctx.Done():
			flushAllIfNecessary()
			pi.releaseAssembler(assembler)

			rotateSession(true)

			if isOnline {
				pi.saveSession(session, "")
			} else {
				pi.saveSession(session, "import process cancelled")
				pi.notificationController.Notify("pcap.canceled", session)
			}

			onComplete(true)

			return
		}
	}
}

func (pi *PcapImporter) saveSession(session *ImportingSession, err string) {
	pi.mSessions.Lock()
	if err == "" {
		session.CompletedAt = time.Now()
	} else {
		session.ImportingError = err
	}
	pi.mSessions.Unlock()

	if _, _err := pi.storage.Insert(ImportingSessions).One(session); _err != nil {
		log.WithError(_err).WithField("session", session).Error("failed to insert importing stats")
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

func (pi *PcapImporter) newSession(isOnline bool) (*ImportingSession, context.Context) {
	ctx, cancelFunc := context.WithCancel(context.Background())
	session := &ImportingSession{
		ID:                NewRowID(),
		StartedAt:         time.Now(),
		PacketsPerService: make(map[uint16]flowCount),
		cancelFunc:        cancelFunc,
	}

	if isOnline {
		pi.mLiveCapture.Lock()
		pi.currentLiveSession = session
		pi.mLiveCapture.Unlock()
	}

	return session, ctx
}

func createNewPcap(handle *pcap.Handle) (*os.File, *pcapgo.Writer) {
	currentFile, err := ioutil.TempFile(ProcessingPcapsBasePath, "live-*.pcap")
	if err != nil {
		log.WithError(err).Panic("failed to create a pcap temp file")
	}
	currentWriter := pcapgo.NewWriter(currentFile)
	currentWriter.WriteFileHeader(snapshotLen, handle.LinkType())

	return currentFile, currentWriter
}

func savePcap(session *ImportingSession, file *os.File) {
	if err := file.Close(); err != nil {
		log.WithError(err).Panic("failed to close live pcap file")
	}
	filePath := file.Name()
	hash, err := Sha256Sum(filePath)
	if err != nil {
		log.WithError(err).Panic("failed to calculate pcap sha256")
	}
	session.Hash = hash
	session.Size = FileSize(filePath)

	moveProcessingFile(session, filepath.Base(filePath))
}

func createSSHClient(sshConfig *SSHConfig) (client *ssh.Client, err error) {
	var authMethod ssh.AuthMethod
	if sshConfig.Password != "" {
		authMethod = ssh.Password(sshConfig.Password)
	} else if sshConfig.PrivateKey != "" {
		var signer ssh.Signer
		if sshConfig.Passphrase != "" {
			signer, err = ssh.ParsePrivateKeyWithPassphrase([]byte(sshConfig.PrivateKey), []byte(sshConfig.Passphrase))
		} else {
			signer, err = ssh.ParsePrivateKey([]byte(sshConfig.PrivateKey))
		}
		if err != nil {
			return nil, err
		}
		authMethod = ssh.PublicKeys(signer)
	} else {
		return nil, errors.New("provide either a password or a passphrase to connect with ssh")
	}
	if sshConfig.Port == 0 {
		sshConfig.Port = 22
	}
	if sshConfig.User == "" {
		sshConfig.User = "root"
	}

	var hostKeyCallback ssh.HostKeyCallback
	if sshConfig.ServerPublicKey != "" {
		publicKey, err := ssh.ParsePublicKey([]byte(sshConfig.ServerPublicKey))
		if err != nil {
			return nil, err
		}
		hostKeyCallback = ssh.FixedHostKey(publicKey)
	} else {
		hostKeyCallback = ssh.InsecureIgnoreHostKey()
	}

	client, err = ssh.Dial("tcp", net.JoinHostPort(sshConfig.Host, fmt.Sprintf("%v", sshConfig.Port)), &ssh.ClientConfig{
		User:            sshConfig.User,
		HostKeyCallback: hostKeyCallback,
		Auth:            []ssh.AuthMethod{authMethod},
	})
	if err != nil {
		return nil, err
	}

	return client, nil
}

func generateBffFilters(captureOptions CaptureOptions) string {
	bffFilter := "tcp"
	if captureOptions.IncludedServices != nil {
		bffFilter += " and port " + strings.Trim(strings.Join(
			strings.Fields(fmt.Sprint(captureOptions.IncludedServices)), " and port "), "[]")
	}

	if captureOptions.ExcludedServices != nil {
		bffFilter += " and not port " + strings.Trim(strings.Join(
			strings.Fields(fmt.Sprint(captureOptions.ExcludedServices)), " and not port "), "[]")
	}

	return bffFilter
}

func deleteProcessingFile(fileName string) {
	if err := os.Remove(ProcessingPcapsBasePath + fileName); err != nil {
		log.WithError(err).Error("failed to delete processing file")
	}
}

func moveProcessingFile(session *ImportingSession, fileName string) {
	if err := os.Rename(ProcessingPcapsBasePath+fileName,
		PcapsBasePath+session.ID.Hex()+path.Ext(fileName)); err != nil {
		log.WithError(err).Error("failed to move processed file")
	}
}
