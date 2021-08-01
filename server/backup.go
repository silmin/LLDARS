package server

import (
	"context"
	"log"
	"net"
	"time"

	"github.com/silmin/lldars/pkg/lldars"
)

func backupRegularly(ctx context.Context, serverId uint32, origin string) {
	ticker := time.NewTicker(time.Duration(BackupIntervalMinute) * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			log.Println("--Backup Objects Regularly--")
			backup(ctx, serverId, origin)
		}
	}
}

func backup(ctx context.Context, serverId uint32, origin string) {
	dcCtx, dcClose := context.WithTimeout(ctx, time.Duration(TimeoutSeconds)*time.Second)
	defer dcClose()

	serviceAddrChan := make(chan string)
	go discoverBroadcast(dcCtx, serverId, serviceAddrChan)

	for {
		select {
		case addr := <-serviceAddrChan:
			log.Printf("service addr: %s\n", addr)
			go backupObjects(addr, serverId, origin)
		case <-dcCtx.Done():
			dcClose()
			return
		}
	}
}

func backupObjects(addr string, serverId uint32, origin string) {
	conn, err := net.Dial("tcp", addr)
	Error(err)
	defer conn.Close()

	ip, _ := lldars.ParseIpPort(conn.LocalAddr().String())
	sl := lldars.NewBackupObjectRequest(serverId, net.ParseIP(ip).To4(), 0)
	conn.Write(sl.Marshal())

	buf := make([]byte, lldars.LLDARSLayerSize)
	l, err := conn.Read(buf)
	Error(err)
	rl := lldars.Unmarshal(buf[:l])

	if rl.Type == lldars.AcceptBackupObject {
		sendObjects(conn, serverId)
	}

	log.Println("--Completed Backup--")
	return
}
