package server

import (
	"fmt"
	"net"
	"time"

	"github.com/silmin/lldars/pkg/lldars"
)

func receiveBackupObjects(conn net.Conn, rl lldars.LLDARSLayer, serverId uint32) {
	defer conn.Close()

	sl := lldars.NewAcceptBackupObject(serverId, localIP(conn), ServicePort)
	msg := sl.Marshal()
	conn.Write(msg)

	path := fmt.Sprintf("%s/%d/", BackupObjectsPath, rl.ServerId)

	receiveObjects(conn, path)
}

func genFilename() string {
	t := time.Now()
	return fmt.Sprintf("%s.zip", t.Format("20060102T150405.000"))
}
