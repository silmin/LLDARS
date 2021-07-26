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
	IntervalSeconds = 1
	TimeoutSeconds  = 10
	ServicePort     = 60001
	SendObjectPath  = "./send_data/"
	SyncObjectPath  = "./sync_data/"
)

func Server(ctx context.Context, bcAddr string, origin string, mode lldars.LLDARSServeMode) {

	serverId := uuid.New()

	if mode == lldars.RevivalMode {
		revCtx, revClose := context.WithTimeout(ctx, time.Duration(TimeoutSeconds)*time.Second)
		defer revClose()

		serviceAddrChan := make(chan string)
		go discoverBroadcast(revCtx, serverId, serviceAddrChan)

		for {
			select {
			case addr := <-serviceAddrChan:
				log.Printf("service addr: %s\n", addr)
				// sync
			case <-revCtx.Done():
				revClose()
				break
			}
		}
	}
	bcCtx, bcClose := context.WithCancel(ctx)
	defer bcClose()

	go listenDiscoverBroadcast(bcCtx, serverId, bcAddr, origin)
	listenService()

	return
}

func listenService() {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", ServicePort))
	Error(err)
	for {
		conn, err := ln.Accept()
		Error(err)
		go handleService(conn)
	}
}

func handleService(conn net.Conn) {
	defer conn.Close()
	buf := make([]byte, lldars.LLDARSLayerSize)
	l, err := conn.Read(buf)
	Error(err)
	msg := buf[:l]
	rl := lldars.Unmarshal(msg)
	log.Printf("Receive from: %v\tmsg: %s\n", rl.Origin, rl.Payload)

	if rl.Type == lldars.GetObjectRequest {
		buf := make([]byte, rl.Length)
		_, err := conn.Read(buf)
		Error(err)

		sendObjects(conn, serverId)
	} else if rl.Type == lldars.SyncObjectRequest {
		buf := make([]byte, rl.Length)
		_, err := conn.Read(buf)
		Error(err)

		receiveSyncObjects(conn, rl, serverId)
	}
	return
}

func listenDiscoverBroadcast(ctx context.Context, serverId uuid.UUID, listenAddr string, origin string) {
	udpAddr, err := net.ResolveUDPAddr("udp", listenAddr)
	Error(err)
	udpLn, err := net.ListenUDP("udp", udpAddr)
	Error(err)
	defer udpLn.Close()

	log.Printf("Listened Delivering Requests *:* udp > %s\n", listenAddr)
	buf := make([]byte, lldars.LLDARSLayerSize)
	for {
		length, err := udpLn.Read(buf)
		Error(err)
		msg := buf[:length]
		rl := lldars.Unmarshal(msg)
		log.Printf("Receive BC from: %v\n", rl.Origin)

		buf = make([]byte, rl.Length)
		_, err = udpLn.Read(buf)
		Error(err)

		if rl.Type == lldars.DiscoverBroadcast {
			sl := lldars.NewServerPortNotify(serverId, net.ParseIP(origin), ServicePort)
			ipp := fmt.Sprintf("%s:%d", rl.Origin.String(), rl.ServicePort)
			ackAddr, err := net.ResolveUDPAddr("udp", ipp)
			Error(err)
			udpLn.WriteToUDP([]byte(sl.Marshal()), ackAddr)
			log.Printf("Ack to: %v\tmsg: %s\n", ackAddr.IP.String(), sl.Payload)
		}
	}
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
