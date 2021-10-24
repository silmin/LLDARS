package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"strconv"
	"time"

	"github.com/silmin/lldars/pkg/lldars"
)

func DiscoverBroadcast(ctx context.Context, serverId uint32, serviceAddrChan chan<- string, bcAddr string) {
	conn, err := net.Dial("udp", bcAddr)
	Error(err)
	defer conn.Close()
	log.Printf("Connected > %s\n", bcAddr)

	ticker := time.NewTicker(time.Duration(IntervalSeconds) * time.Second)
	defer ticker.Stop()

	// listen udp for ack
	ip, _ := lldars.ParseIpPort(conn.LocalAddr().String())
	ln, err := net.Listen("tcp", fmt.Sprintf("%v:0", ip))
	Error(err)

	listenCtx, listenCancel := context.WithTimeout(ctx, time.Duration(TimeoutSeconds)*time.Second)
	defer listenCancel()
	go listenAck(listenCtx, ln, serviceAddrChan)

	addr, port := lldars.ParseIpPort(ln.Addr().String())
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
			log.Printf("Cast > %v : “%s”\n", bcAddr, l.Payload)
		}
	}
}

func listenAck(ctx context.Context, ln net.Listener, serviceAddrChan chan<- string) {
	log.Println("Start listenAck()")
	for {
		conn, err := ln.Accept()
		Error(err)
		go handleAck(ctx, conn, serviceAddrChan)
	}
}

func handleAck(ctx context.Context, conn net.Conn, serviceAddrChan chan<- string) {
	defer conn.Close()
	rl := ReadLLDARSHeader(conn)
	if !lldars.IsEqualIP(rl.Origin, localConnIP(conn)) {
		rl.Payload = ReadLLDARSPayload(conn, rl.Length)
		log.Printf("Receive Ack from: %v\tsId: %v\tmsg: %s\n", rl.Origin, rl.ServerId, rl.Payload)
		serviceAddrChan <- fmt.Sprintf("%s:%d", rl.Origin.String(), rl.ServicePort)
	}
}
