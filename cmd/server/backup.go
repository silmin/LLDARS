package main

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

func (s Server) backupRegularly(ctx context.Context) {
	ticker := time.NewTicker(time.Duration(BackupIntervalMinute) * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			log.Println("--Backup Objects Regularly--")
			backupObjects(ctx, s.Id, s.ServerBCAddr)
		}
	}
}

func backupObjects(ctx context.Context, serverId uint32, bcAddr string) {
	dcCtx, dcClose := context.WithTimeout(ctx, time.Duration(BackupBCTimeoutSeconds)*time.Second)
	defer dcClose()

	serviceAddrChan := make(chan string)
	go DiscoverBroadcast(dcCtx, serverId, serviceAddrChan, bcAddr)

	wg := new(sync.WaitGroup)

	for {
		select {
		case addr := <-serviceAddrChan:
			log.Printf("service addr: %s\n", addr)
			wg.Add(1)
			go handleBackup(wg, addr, serverId)
		case <-dcCtx.Done():
			wg.Wait()
			log.Println("--End Backup Objects--")
			return
		}
	}
}

func handleBackup(wg *sync.WaitGroup, addr string, serverId uint32) {
	conn, err := net.Dial("tcp", addr)
	Error(err)
	defer func() {
		conn.Close()
		wg.Done()
		log.Printf("-Completed Backup to %v-", addr)
	}()

	ip := localConnIP(conn)
	sl := lldars.NewBackupObjectRequest(serverId, ip, 0)
	conn.Write(sl.Marshal())

	rl := ReadLLDARSHeader(conn)

	if rl.Type == lldars.AcceptBackupObject {
		sendObjects(conn, serverId, LLDARSObjectPath)
	} else if rl.Type == lldars.RejectBackupObject {
		log.Println("-Rejected Backup-")
	}

	return
}

func receiveBackupObjects(conn net.Conn, rl lldars.LLDARSLayer, serverId uint32, cache *IdCache) {
	defer conn.Close()

	if rl.ServerId != 0 && cache.Exists(cacheBackupKey(rl.ServerId)) {
		sl := lldars.NewRejectBackupObject(serverId, localConnIP(conn), lldars.ServicePort)
		conn.Write(sl.Marshal())
		return
	} else {
		sl := lldars.NewAcceptBackupObject(serverId, localConnIP(conn), lldars.ServicePort)
		conn.Write(sl.Marshal())
		cache.Push(cacheBackupKey(rl.ServerId), rl.ServerId, time.Now().Add(ExpirationSecondsOfBackup*time.Second).UnixNano())
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

func cacheBackupKey(serverId uint32) string {
	return fmt.Sprintf("%v-backup", serverId)
}
