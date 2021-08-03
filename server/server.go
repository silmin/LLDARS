package server

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/google/uuid"
	"github.com/silmin/lldars/pkg/lldars"
)

const (
	IntervalSeconds        = 1
	TimeoutSeconds         = 10
	ExpirationSecondsOfAck = 30
	ServicePort            = 60001
	LLDARSObjectPath       = "./send_data/"
	BackupObjectsPath      = "./backups/"
	BroadcastAddr          = "192.168.100.255:60000"
)

func Server(ctx context.Context, bcAddr string, origin string, id uint32, mode lldars.LLDARSServeMode) {
	var serverId uint32
	if id != 0 {
		serverId = id
	} else {
		serverId = uuid.New().ID()
	}
	log.Printf("Server ID: %v\n", serverId)

	if mode == lldars.RevivalMode {
		syCtx, syClose := context.WithCancel(ctx)
		defer syClose()
		syncObjects(syCtx, serverId)
	}
	bcCtx, bcClose := context.WithCancel(ctx)
	defer bcClose()
	brCtx, brClose := context.WithCancel(ctx)
	defer brClose()

	ackCache := NewIdCache()
	backupCache := NewIdCache()

	go listenDiscoverBroadcast(bcCtx, serverId, bcAddr, origin, ackCache)
	go backupRegularly(brCtx, serverId, origin)
	listenService(serverId, backupCache)

	return
}

func listenService(serverId uint32, cache *IdCache) {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", ServicePort))
	Error(err)
	for {
		conn, err := ln.Accept()
		Error(err)
		go handleService(conn, serverId, cache)
	}
}

func handleService(conn net.Conn, serverId uint32, cache *IdCache) {
	defer conn.Close()
	rl := ReadLLDARSHeader(conn)
	log.Printf("Receive from: %v\n", rl.Origin)

	if rl.Type == lldars.GetObjectRequest {
		_ = ReadLLDARSPayload(conn, rl.Length)

		sendObjects(conn, serverId, LLDARSObjectPath)
	} else if rl.Type == lldars.SyncObjectRequest {
		_ = ReadLLDARSPayload(conn, rl.Length)

		sendSyncObjects(conn, rl, serverId)
	} else if rl.Type == lldars.BackupObjectRequest {
		_ = ReadLLDARSPayload(conn, rl.Length)

		receiveBackupObjects(conn, rl, serverId, cache)
	}

	return
}

func listenDiscoverBroadcast(ctx context.Context, serverId uint32, listenAddr string, origin string, cache *IdCache) {
	udpAddr, err := net.ResolveUDPAddr("udp", listenAddr)
	Error(err)
	udpLn, err := net.ListenUDP("udp", udpAddr)
	Error(err)
	defer udpLn.Close()

	log.Printf("Listened Delivering Requests *:* udp > %s\n", listenAddr)
	for {
		buf := make([]byte, lldars.LLDARSLayerSize+len(lldars.DiscoverBroadcastPayload))
		l, err := udpLn.Read(buf)
		Error(err)
		msg := buf[:l]
		rl := lldars.Unmarshal(msg)
		log.Printf("Receive BC from: %v\n", rl.Origin)

		if rl.Type == lldars.DiscoverBroadcast && rl.Origin.String() != origin {
			if rl.ServerId != 0 {
				if !cache.Exists(cacheAckKey(serverId)) {
					cache.Push(cacheAckKey(serverId), rl.ServerId, time.Now().Add(ExpirationSecondsOfAck*time.Second).UnixNano())
					ackBroadcast(serverId, rl, origin)
				}
			} else {
				ackBroadcast(serverId, rl, origin)
			}
		}
	}
}

func ackBroadcast(serverId uint32, rl lldars.LLDARSLayer, origin string) {
	sl := lldars.NewServerPortNotify(serverId, net.ParseIP(origin), ServicePort)
	ipp := fmt.Sprintf("%s:%d", rl.Origin.String(), rl.ServicePort)
	conn, err := net.Dial("tcp", ipp)
	Error(err)
	defer conn.Close()
	conn.Write(sl.Marshal())
	log.Printf("Ack to: %v\tmsg: %s\n", conn.RemoteAddr().String(), sl.Payload)
}

func localConnIP(conn net.Conn) net.IP {
	ipstr, _ := lldars.ParseIpPort(conn.LocalAddr().String())
	return net.ParseIP(ipstr).To4()
}

func cacheAckKey(serverId uint32) string {
	return fmt.Sprintf("%v-ack", serverId)
}

func Error(_err error) {
	if _err != nil {
		log.Panic(_err)
	}
}
