package main

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
	go DiscoverBroadcast(dcCtx, serverId, serviceAddrChan)

	wg := new(sync.WaitGroup)

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

func handleSync(wg *sync.WaitGroup, addr string, serverId uint32) {
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

	rl := ReadLLDARSHeader(conn)

	if rl.Type == lldars.AcceptSyncObject {
		_ = ReadLLDARSPayload(conn, rl.Length)
		receiveObjects(conn, LLDARSObjectPath)
	} else if rl.Type == lldars.RejectSyncObject {
		log.Println("-Rejected Sync-")
	}
	return
}

func sendSyncObjects(conn net.Conn, rl lldars.LLDARSLayer, serverId uint32) {
	if !hasBackup(rl.ServerId) {
		sl := lldars.NewRejectSyncObject(serverId, localConnIP(conn), ServicePort)
		conn.Write(sl.Marshal())
		return
	}
	sl := lldars.NewAcceptSyncObject(serverId, localConnIP(conn), ServicePort)
	conn.Write(sl.Marshal())

	sendObjects(conn, serverId, getBackupDirPath(rl.ServerId))
}
