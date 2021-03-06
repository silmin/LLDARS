package main

import (
	"io/ioutil"
	"log"
	"net"
	"path/filepath"
	"time"

	"github.com/yargevad/filepathx"

	"github.com/silmin/lldars/pkg/lldars"
)

func sendObjects(conn net.Conn, serverId uint32, path string) {
	log.Printf("serverId: %d", serverId)

	paths := getObjectPaths(path)
	for _, path := range paths {
		obj, err := ioutil.ReadFile(path)
		Error(err)
		sl := lldars.NewDeliveryObject(serverId, localConnIP(conn), lldars.ServicePort, obj)
		conn.Write(sl.Marshal())
		log.Printf("Send Object > %s len: %d\n", conn.RemoteAddr().String(), sl.Length)
	}

	sl := lldars.NewEndOfDelivery(serverId, localConnIP(conn), lldars.ServicePort)
	conn.Write(sl.Marshal())

	time.Sleep(time.Second)
	return
}

func getObjectPaths(path string) []string {
	pat := filepath.Join(path, "/**/*.zip")
	files, err := filepathx.Glob(pat)
	Error(err)
	return files
}
