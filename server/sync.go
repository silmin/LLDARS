package server

import (
	"context"
	"log"
	"net"
	"sync"
	"time"

	"github.com/silmin/lldars/pkg/lldars"
)

const (
	SyncBCTimeoutSeconds = 10
)

func syncObjects(ctx context.Context, serverId uint32) {
	log.Println("--Start Sync Objects--")

	dcCtx, dcClose := context.WithTimeout(ctx, time.Duration(SyncBCTimeoutSeconds)*time.Second)
	defer dcClose()

	serviceAddrChan := make(chan string)
	go discoverBroadcast(dcCtx, serverId, serviceAddrChan)

	var wg sync.WaitGroup

	for {
		select {
		case addr := <-serviceAddrChan:
			log.Printf("service addr: %s\n", addr)
			wg.Add(1)
			go handleSync(wg, addr, serverId)
		case <-dcCtx.Done():
			wg.Wait()
			log.Println("--End Sync Objects--")
			return
		}
	}
}

func handleSync(wg sync.WaitGroup, addr string, serverId uint32) {
	conn, err := net.Dial("tcp", addr)
	Error(err)
	defer func() {
		conn.Close()
		wg.Done()
		log.Printf("-Completed Sync from %s-", addr)
	}()

	ip, _ := lldars.ParseIpPort(conn.LocalAddr().String())
	sl := lldars.NewSyncObjectRequest(serverId, net.ParseIP(ip).To4(), 0)
	conn.Write(sl.Marshal())

	receiveObjects(conn, LLDARSObjectPath)

	return
}
