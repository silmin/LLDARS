package server

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"

	"github.com/google/uuid"
	"github.com/silmin/lldars/pkg/lldars"
)

const (
	IntervalSeconds   = 1
	TimeoutSeconds    = 10
	ServicePort       = 60001
	LLDARSObjectPath  = "./send_data"
	BackupObjectsPath = "./backups"
	BroadcastAddr     = "192.168.100.255:60000"
)

func Server(ctx context.Context, bcAddr string, origin string, mode lldars.LLDARSServeMode) {
	serverId := uuid.New().ID()

	if mode == lldars.RevivalMode {
		syCtx, syClose := context.WithCancel(ctx)
		defer syClose()
		syncObjects(syCtx, serverId)
	}
	bcCtx, bcClose := context.WithCancel(ctx)
	defer bcClose()
	brCtx, brClose := context.WithCancel(ctx)
	defer brClose()

	go listenDiscoverBroadcast(bcCtx, serverId, bcAddr, origin)
	go backupRegularly(brCtx, serverId, origin)
	listenService(serverId)

	return
}

func listenService(serverId uint32) {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", ServicePort))
	Error(err)
	for {
		conn, err := ln.Accept()
		Error(err)
		go handleService(conn, serverId)
	}
}

func handleService(conn net.Conn, serverId uint32) {
	defer conn.Close()
	buf := make([]byte, lldars.LLDARSLayerSize)
	l, err := conn.Read(buf)
	Error(err)
	msg := buf[:l]
	rl := lldars.Unmarshal(msg)
	log.Printf("Receive from: %v\n", rl.Origin)

	if rl.Type == lldars.GetObjectRequest {
		buf := make([]byte, rl.Length)
		_, err := conn.Read(buf)
		Error(err)

		sendObjects(conn, serverId)
	} else if rl.Type == lldars.SyncObjectRequest {
		buf := make([]byte, rl.Length)
		_, err := conn.Read(buf)
		Error(err)

		sendObjects(conn, serverId)
	} else if rl.Type == lldars.BackupObjectRequest {
		buf := make([]byte, rl.Length)
		_, err := conn.Read(buf)
		Error(err)

		receiveBackupObjects(conn, rl, serverId)
	}

	return
}

func listenDiscoverBroadcast(ctx context.Context, serverId uint32, listenAddr string, origin string) {
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

		if rl.Type == lldars.DiscoverBroadcast {
			if (rl.ServerId == 0 || hasBackup(rl.ServerId)) && rl.Origin.String() != origin {
				ackBroadcast(serverId, rl, udpLn, origin)
			}
		}
	}
}

func ackBroadcast(serverId uint32, rl lldars.LLDARSLayer, udpLn *net.UDPConn, origin string) {
	sl := lldars.NewServerPortNotify(serverId, net.ParseIP(origin), ServicePort)
	ipp := fmt.Sprintf("%s:%d", rl.Origin.String(), rl.ServicePort)
	ackAddr, err := net.ResolveUDPAddr("udp", ipp)
	Error(err)
	udpLn.WriteToUDP([]byte(sl.Marshal()), ackAddr)
	log.Printf("Ack to: %v\tmsg: %s\n", ackAddr.IP.String(), sl.Payload)
}

func hasBackup(serverId uint32) bool {
	path := fmt.Sprintf("%s/%d", BackupObjectsPath, serverId)
	f, err := os.Stat(path)
	return f.IsDir() && !os.IsNotExist(err)
}

func localIP(conn net.Conn) net.IP {
	ipstr, _ := lldars.ParseIpPort(conn.LocalAddr().String())
	return net.ParseIP(ipstr).To4()
}

func Error(_err error) {
	if _err != nil {
		log.Panic(_err)
	}
}
