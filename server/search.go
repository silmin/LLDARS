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
	addr, _ := lldars.ParseIpPort(udpLn.LocalAddr().String())
	localAddr := net.ParseIP(addr).To4()

	for {
		buf := make([]byte, lldars.LLDARSLayerSize)
		l, err := udpLn.Read(buf)
		Error(err)
		rl := lldars.Unmarshal(buf[:l])

		buf = make([]byte, rl.Length)
		l, err = udpLn.Read(buf)
		Error(err)
		rl.Payload = buf[:l]

		if !lldars.IsEqualIP(rl.Origin, localAddr) {
			log.Printf("Recieve from: %v\tmsg: %s\n", rl.Origin, rl.Payload)
			serviceAddrChan <- fmt.Sprintf("%s:%d", rl.Origin.String(), rl.ServicePort)
		}
	}
}
