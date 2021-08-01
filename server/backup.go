package server

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"sync"
	"time"

	"github.com/silmin/lldars/pkg/lldars"
)

const (
	BackupIntervalMinute   = 1
	BackupBCTimeoutSeconds = 10
)

func backupRegularly(ctx context.Context, serverId uint32, origin string) {
	ticker := time.NewTicker(time.Duration(BackupIntervalMinute) * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			log.Println("--Backup Objects Regularly--")
			backupObjects(ctx, serverId, origin)
		}
	}
}

func backupObjects(ctx context.Context, serverId uint32, origin string) {
	dcCtx, dcClose := context.WithTimeout(ctx, time.Duration(BackupBCTimeoutSeconds)*time.Second)
	defer dcClose()

	serviceAddrChan := make(chan string)
	go discoverBroadcast(dcCtx, serverId, serviceAddrChan)

	var wg sync.WaitGroup

	for {
		select {
		case addr := <-serviceAddrChan:
			log.Printf("service addr: %s\n", addr)
			wg.Add(1)
			go handleBackup(wg, addr, serverId, origin)
		case <-dcCtx.Done():
			wg.Wait()
			log.Println("--End Backup Objects--")
			return
		}
	}
}

func handleBackup(wg sync.WaitGroup, addr string, serverId uint32, origin string) {
	conn, err := net.Dial("tcp", addr)
	Error(err)
	defer func() {
		conn.Close()
		wg.Done()
		log.Printf("-Completed Backup to %v-", addr)
	}()

	ip, _ := lldars.ParseIpPort(conn.LocalAddr().String())
	sl := lldars.NewBackupObjectRequest(serverId, net.ParseIP(ip).To4(), 0)
	conn.Write(sl.Marshal())

	buf := make([]byte, lldars.LLDARSLayerSize)
	l, err := conn.Read(buf)
	Error(err)
	rl := lldars.Unmarshal(buf[:l])

	if rl.Type == lldars.AcceptBackupObject {
		sendObjects(conn, serverId, LLDARSObjectPath)
	} else if rl.Type == lldars.RejectBackupObject {
		log.Println("-Rejected Backup-")
	}

	return
}

func receiveBackupObjects(conn net.Conn, rl lldars.LLDARSLayer, serverId uint32) {
	defer conn.Close()

	sl := lldars.NewAcceptBackupObject(serverId, localIP(conn), ServicePort)
	msg := sl.Marshal()
	conn.Write(msg)

	path := getBackupPath(rl.ServerId)
	receiveObjects(conn, path)
}

func hasBackup(serverId uint32) bool {
	f, err := os.Stat(getBackupPath(serverId))
	return err == nil && f.IsDir()
}

func getBackupPath(serverId uint32) string {
	return fmt.Sprintf("%s/%v", BackupObjectsPath, serverId)
}
