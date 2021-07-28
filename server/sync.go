package server

import (
	"context"
	"fmt"
	"log"
	"net"
	"strconv"
	"time"

	"github.com/silmin/lldars/pkg/lldars"
)

const (
	BroadcastAddr = "192.168.100.255:60000"
)

func sync(ctx context.Context, serverId uint32) {
	revCtx, revClose := context.WithTimeout(ctx, time.Duration(TimeoutSeconds)*time.Second)
	defer revClose()

	serviceAddrChan := make(chan string)
	go discoverBroadcast(revCtx, serverId, serviceAddrChan)

	for {
		select {
		case addr := <-serviceAddrChan:
			log.Printf("service addr: %s\n", addr)
			syncObjects(addr, serverId)
		case <-revCtx.Done():
			revClose()
			break
		}
	}
}

func syncObjects(addr string, serverId uint32) {
	conn, err := net.Dial("tcp", addr)
	Error(err)
	defer conn.Close()

	ip, _ := lldars.ParseIpPort(conn.LocalAddr().String())
	log.Printf("conn.LocalAddr: %s", conn.LocalAddr().String())
	sl := lldars.NewSyncObjectRequest(0, net.ParseIP(ip).To4(), 0)
	msg := sl.Marshal()
	conn.Write(msg)

	receiveObjects(conn, LLDARSObjectPath)

	log.Println("End getObjects()")
	return
}

func discoverBroadcast(ctx context.Context, serverId uint32, servicePortChan chan<- string) {
	conn, err := net.Dial("udp", BroadcastAddr)
	Error(err)
	defer conn.Close()
	log.Printf("Connected > %s\n", BroadcastAddr)

	ticker := time.NewTicker(time.Duration(IntervalSeconds) * time.Second)
	defer ticker.Stop()

	// listen udp for ack
	ip, _ := lldars.ParseIpPort(conn.LocalAddr().String())
	udpAddr := &net.UDPAddr{
		IP:   net.ParseIP(ip),
		Port: 0,
	}
	udpLn, err := net.ListenUDP("udp", udpAddr)

	listenCtx, listenCancel := context.WithTimeout(ctx, time.Duration(TimeoutSeconds)*time.Second)
	defer listenCancel()
	go listenAck(listenCtx, udpLn, servicePortChan)

	addr, port := lldars.ParseIpPort(udpLn.LocalAddr().String())
	Error(err)
	p, _ := strconv.Atoi(port)
	l := lldars.NewDiscoverBroadcast(serverId, net.ParseIP(addr).To4(), uint16(p))
	msg := l.Marshal()
	for {
		select {
		case <-ctx.Done():
			log.Println("End Broadcast")
			return
		case <-ticker.C:
			// broadcast
			conn.Write(msg)
			log.Printf("Cast > %v : “%s”\n", BroadcastAddr, l.Payload)
		}
	}
}

func listenAck(ctx context.Context, udpLn *net.UDPConn, serviceAddrChan chan<- string) {
	log.Println("Start listenAck()")

	for {
		buf := make([]byte, lldars.LLDARSLayerSize)
		l, err := udpLn.Read(buf)
		Error(err)
		msg := string(buf[:l])
		rl := lldars.Unmarshal([]byte(msg))

		buf = make([]byte, rl.Length)
		l, err = udpLn.Read(buf)
		Error(err)
		msg = string(buf[:l])

		log.Printf("Recieve from: %v\tmsg: %s\n", rl.Origin, msg)

		serviceAddrChan <- fmt.Sprintf("%s:%d", rl.Origin.String(), rl.ServicePort)
	}
}
