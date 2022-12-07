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
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/eciavatta/caronte/pkg/tcpassembly"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"github.com/google/gopacket/pcapgo"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
)

const initialAssemblerPoolSize = 16
const initialSessionRotationInterval = 2 * time.Minute
const snapshotLen = 1024
const remoteCapturePipeName = "remote-pipe.pcap"

var PcapsBasePath string
var ProcessingPcapsBasePath string
var ConnectionPcapsBasePath string

func init() {
	exe, err := os.Executable()
	if err != nil {
		panic(err)
	}

	PcapsBasePath = filepath.Join(filepath.Dir(exe), "pcaps")
	ProcessingPcapsBasePath = filepath.Join(PcapsBasePath, "processing")
	ConnectionPcapsBasePath = filepath.Join(PcapsBasePath, "connections")
}

type PcapImporter struct {
	storage                 Storage
	streamPool              *tcpassembly.StreamPool
	assemblers              []*tcpassembly.Assembler
	sessions                map[RowID]*ImportingSession
	mAssemblers             sync.Mutex
	mSessions               sync.Mutex
	serverNet               *net.IPNet
	notificationController  NotificationController
	liveCaptureHandle       *pcap.Handle
	liveCaptureType         string
	mLiveCapture            sync.Mutex
	currentLiveSession      *ImportingSession
	sessionRotationInterval time.Duration
	packetsStatusChannel    chan bool // true for processed packets, false for invalid packets
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

type PacketsStatistics struct {
	ProcessedPackets uint64 `json:"processed_packets"`
	InvalidPackets   uint64 `json:"invalid_packets"`
	PacketsPerMinute uint64 `json:"packets_per_minute"`
}

type flowCount [2]int

func NewPcapImporter(storage Storage, serverNet *net.IPNet, rulesManager RulesManager, notificationController NotificationController) *PcapImporter {
	streamPool := tcpassembly.NewStreamPool(NewBiDirectionalStreamFactory(storage, serverNet, rulesManager, notificationController))

	if err := os.MkdirAll(ProcessingPcapsBasePath, 0755); err != nil {
		log.WithError(err).Panic("failed to create processing pcaps folder")
	}

	if err := os.MkdirAll(ConnectionPcapsBasePath, 0755); err != nil {
		log.WithError(err).Panic("failed to create connections pcaps folder")
	}

	var result []ImportingSession
	if err := storage.Find(ImportingSessions).All(&result); err != nil {
		log.WithError(err).Panic("failed to retrieve importing sessions")
	}

	sessions := make(map[RowID]*ImportingSession)
	for i := 0; i < len(result); i++ {
		session := result[i]
		sessions[session.ID] = &session
	}

	pcapImporter := &PcapImporter{
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
		packetsStatusChannel:    make(chan bool),
	}
	go pcapImporter.notificationService()

	return pcapImporter
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

	processingPath := filepath.Join(ProcessingPcapsBasePath, fileName)

	hash, err := Sha256Sum(processingPath)
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

	handle, err := pcap.OpenOffline(processingPath)
	if err != nil {
		pi.mSessions.Unlock()
		deleteProcessingFile(fileName)
		log.WithError(err).Panic("failed to process pcap")
	}

	session, ctx := pi.newSession(false)
	session.Hash = hash
	session.Size = FileSize(processingPath)

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

func (pi *PcapImporter) StartLocalCapture(captureOptions CaptureOptions) error {
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
	pi.liveCaptureType = "local"
	go pi.handle(handle, nil, true, nil, func(_ bool) {
		handle.Close()
	})

	return nil
}

func (pi *PcapImporter) StopCapture() error {
	pi.mLiveCapture.Lock()
	defer pi.mLiveCapture.Unlock()

	if pi.liveCaptureHandle == nil {
		return errors.New("live capture is already stopped")
	}

	pi.currentLiveSession.cancelFunc()
	pi.liveCaptureHandle = nil
	pi.liveCaptureType = ""

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

func (pi *PcapImporter) StartRemoteCapture(sshConfig SSHConfig, captureOptions CaptureOptions) error {
	pi.mLiveCapture.Lock()
	defer pi.mLiveCapture.Unlock()

	if pi.liveCaptureHandle != nil {
		return errors.New("live capture is already in progress")
	}

	client, err := createSSHClient(&sshConfig)
	if err != nil {
		return err
	}

	// check if tcpdump is present
	checkSession, err := client.NewSession()
	if err != nil {
		return err
	}
	if err = checkSession.Run("tcpdump --version"); err != nil {
		return err
	}
	if err = checkSession.Close(); err != nil && err != io.EOF {
		log.WithError(err).Panic("failed to close ssh session")
	}

	if interfaces, err := pi.ListRemoteInterfaces(sshConfig); err != nil {
		return err
	} else {
		validInterface := false
		for _, ifh := range interfaces {
			if ifh == captureOptions.Interface {
				validInterface = true
			}
		}

		if !validInterface {
			return errors.New("interface not present on remote host")
		}
	}

	pipeName := filepath.Join(ProcessingPcapsBasePath, remoteCapturePipeName)
	if FileExists(pipeName) {
		deleteProcessingFile(remoteCapturePipeName)
	}

	if err = syscall.Mkfifo(pipeName, 0666); err != nil {
		log.WithError(err).Panic("failed to create named pipe")
	}

	namedPipe, err := os.OpenFile(pipeName, os.O_RDWR|os.O_CREATE, 0777)
	if err != nil {
		log.WithError(err).Panic("failed to open named pipe")
	}

	dumpSession, err := client.NewSession()
	if err != nil {
		return err
	}

	dumpSession.Stdout = bufio.NewWriter(namedPipe)
	captureOptions.ExcludedServices = append(captureOptions.ExcludedServices, sshConfig.Port)

	if err := dumpSession.Start(fmt.Sprintf("tcpdump -U -n -w - -i %s %s", captureOptions.Interface, generateBffFilters(captureOptions))); err != nil {
		return fmt.Errorf("failed to start tcpdump on remote host: %s", err)
	}

	handle, err := pcap.OpenOfflineFile(namedPipe)
	if err != nil {
		return fmt.Errorf("failed to open local named pipe: %s", err)
	}

	pi.liveCaptureHandle = handle
	pi.liveCaptureType = "remote"

	go pi.handle(handle, nil, true, nil, func(canceled bool) {
		if killSession, err := client.NewSession(); err == nil {
			if err := killSession.Run("kill -9 $(pidof tcpdump)"); err != nil {
				log.WithError(err).Error("failed to kill tcpdump on remote host")
			}

			if err := killSession.Close(); err != nil && err != io.EOF {
				log.WithError(err).Error("failed to close kill ssh session")
			}

			if err := dumpSession.Close(); err != nil && err != io.EOF {
				log.WithError(err).Error("failed to close dump ssh session")
			}
		} else {
			log.WithError(err).Error("failed to start kill ssh session")
		}

		if err := namedPipe.Close(); err != nil {
			log.WithError(err).Panic("failed to close named pipe")
		}

		deleteProcessingFile(remoteCapturePipeName)

		// TODO: deadlock here
		// handle.Close()
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

// returns "" when stopped, "local" or "remote" when running
func (pi *PcapImporter) GetLiveCaptureStatus() string {
	pi.mLiveCapture.Lock()
	defer pi.mLiveCapture.Unlock()

	return pi.liveCaptureType
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

func (pi *PcapImporter) handle(handle *pcap.Handle, initialSession *ImportingSession, flushAll bool, ctx context.Context, onComplete func(canceled bool)) {
	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
	packetSource.NoCopy = true
	packetSource.Lazy = true
	assembler := pi.takeAssembler()
	packets := packetSource.Packets()

	var currentFile *os.File
	var currentWriter *pcapgo.Writer

	sessionRotationInterval := time.NewTicker(pi.sessionRotationInterval)
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

			if isOnline {
				currentWriter.WritePacket(packet.Metadata().CaptureInfo, packet.Data())
			}

			offlineLock()
			session.ProcessedPackets++
			pi.packetsStatusChannel <- true

			if packet.NetworkLayer() == nil || packet.TransportLayer() == nil ||
				packet.TransportLayer().LayerType() != layers.LayerTypeTCP { // invalid packet
				session.InvalidPackets++
				pi.packetsStatusChannel <- false
				offlineUnlock()

				continue
			}

			tcp := packet.TransportLayer().(*layers.TCP)
			var servicePort uint16
			var index int

			if pi.serverNet != nil {
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
					pi.packetsStatusChannel <- true
					offlineUnlock()

					continue
				}

				fCount, isPresent := session.PacketsPerService[servicePort]
				if !isPresent {
					fCount = flowCount{0, 0}
				}
				fCount[index]++
				session.PacketsPerService[servicePort] = fCount
			}

			offlineUnlock()
			assembler.AssembleWithTimestamp(packet.NetworkLayer().NetworkFlow(), tcp, packet.Metadata().Timestamp, packet, handle.LinkType())
		case <-sessionRotationInterval.C:
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

func (pi *PcapImporter) notificationService() {
	var stats, lastStats PacketsStatistics
	var packetsPerMinute uint64
	ticker := time.NewTicker(3 * time.Second)
	perMinuteTicker := time.NewTicker(time.Minute)

	updateStatistics := func() {
		lastStats = stats
		pi.notificationController.Notify("packets.statistics", stats)
	}

	for {
		select {
		case processed := <-pi.packetsStatusChannel:
			if processed {
				stats.ProcessedPackets++
			} else {
				stats.InvalidPackets++
			}
			packetsPerMinute++
		case <-ticker.C:
			if lastStats != stats {
				updateStatistics()
			}
		case <-perMinuteTicker.C:
			stats.PacketsPerMinute = packetsPerMinute
			packetsPerMinute = 0
		}
	}
}

func createNewPcap(handle *pcap.Handle) (*os.File, *pcapgo.Writer) {
	currentFile, err := os.CreateTemp(ProcessingPcapsBasePath, "live-*.pcap")
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
		Timeout:         30 * time.Second,
	})
	if err != nil {
		return nil, err
	}

	return client, nil
}

func generateBffFilters(captureOptions CaptureOptions) string {
	bffFilter := "tcp"
	if len(captureOptions.IncludedServices) > 0 {
		bffFilter += " and port " + strings.Trim(strings.Join(
			strings.Fields(fmt.Sprint(captureOptions.IncludedServices)), " and port "), "[]")
	}

	if len(captureOptions.ExcludedServices) > 0 {
		bffFilter += " and not port " + strings.Trim(strings.Join(
			strings.Fields(fmt.Sprint(captureOptions.ExcludedServices)), " and not port "), "[]")
	}

	return bffFilter
}

func deleteProcessingFile(fileName string) {
	if err := os.Remove(filepath.Join(ProcessingPcapsBasePath, fileName)); err != nil {
		log.WithError(err).Error("failed to delete processing file")
	}
}

func moveProcessingFile(session *ImportingSession, fileName string) {
	if err := os.Rename(
		filepath.Join(ProcessingPcapsBasePath, fileName),
		filepath.Join(PcapsBasePath, session.ID.Hex()+path.Ext(fileName)),
	); err != nil {
		log.WithError(err).Error("failed to move processed file")
	}
}
