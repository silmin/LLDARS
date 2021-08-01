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
	BackupIntervalMinute      = 1
	BackupBCTimeoutSeconds    = 10
	ExpirationSecondsOfBackup = 60
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

	wg := new(sync.WaitGroup)

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

func handleBackup(wg *sync.WaitGroup, addr string, serverId uint32, origin string) {
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

func receiveBackupObjects(conn net.Conn, rl lldars.LLDARSLayer, serverId uint32, cache *IdCache) {
	defer conn.Close()

	if cache.Exists(rl.ServerId) {
		sl := lldars.NewRejectBackupObject(serverId, localIP(conn), ServicePort)
		conn.Write(sl.Marshal())
		return
	} else {
		sl := lldars.NewAcceptBackupObject(serverId, localIP(conn), ServicePort)
		conn.Write(sl.Marshal())
		cache.Put(rl.ServerId, time.Now().Add(ExpirationSecondsOfBackup*time.Second).UnixNano())
	}

	path := getBackupDirPath(rl.ServerId)
	if !hasBackup(rl.ServerId) {
		err := os.Mkdir(path, 0755)
		Error(err)
	}
	receiveObjects(conn, path)
}

func hasBackup(serverId uint32) bool {
	f, err := os.Stat(getBackupDirPath(serverId))
	return err == nil && f.IsDir()
}

func getBackupDirPath(serverId uint32) string {
	return fmt.Sprintf("%s%v/", BackupObjectsPath, serverId)
}
