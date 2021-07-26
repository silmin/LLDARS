package server

import (
	"io/ioutil"
	"log"
	"net"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/silmin/lldars/pkg/lldars"
)

func sendObjects(conn net.Conn, serverId uuid.UUID) {
	defer conn.Close()

	paths := getObjectPaths(SendObjectPath)
	for _, path := range paths {
		obj, err := ioutil.ReadFile(path)
		Error(err)
		sl := lldars.NewDeliveryObject(serverId, localIP(conn), ServicePort, obj)
		msg := sl.Marshal()
		conn.Write(msg)
		log.Printf("Send Object > %s len: %d\n", conn.RemoteAddr().String(), sl.Length)
	}

	sl := lldars.NewEndOfDelivery(serverId, localIP(conn), ServicePort)
	msg := sl.Marshal()
	conn.Write(msg)
	return
}

func getObjectPaths(path string) []string {
	pat := path + "*.zip"
	files, err := filepath.Glob(pat)
	Error(err)
	return files
}
