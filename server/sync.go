package server

import (
	"context"
	"log"
	"net"
	"time"

	"github.com/silmin/lldars/pkg/lldars"
)

const (
	BroadcastAddr = "192.168.100.255:60000"
)

func sync(ctx context.Context, serverId uint32) {
	log.Println("--Start Sync Objects--")

	dcCtx, dcClose := context.WithTimeout(ctx, time.Duration(TimeoutSeconds)*time.Second)
	defer dcClose()

	serviceAddrChan := make(chan string)
	go discoverBroadcast(dcCtx, serverId, serviceAddrChan)

	for {
		select {
		case addr := <-serviceAddrChan:
			log.Printf("service addr: %s\n", addr)
			go syncObjects(addr, serverId)
		case <-dcCtx.Done():
			dcClose()
			return
		}
	}
}

func syncObjects(addr string, serverId uint32) {
	conn, err := net.Dial("tcp", addr)
	Error(err)
	defer conn.Close()

	ip, _ := lldars.ParseIpPort(conn.LocalAddr().String())
	sl := lldars.NewSyncObjectRequest(0, net.ParseIP(ip).To4(), 0)
	conn.Write(sl.Marshal())

	receiveObjects(conn, LLDARSObjectPath)

	log.Println("--Completed Sync--")
	return
}
