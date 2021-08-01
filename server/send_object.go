package server

import (
	"io/ioutil"
	"log"
	"net"
	"path/filepath"

	"github.com/silmin/lldars/pkg/lldars"
)

func sendObjects(conn net.Conn, serverId uint32, path string) {
	defer conn.Close()
	log.Printf("serverId: %d", serverId)

	paths := getObjectPaths(path)
	for _, path := range paths {
		obj, err := ioutil.ReadFile(path)
		Error(err)
		sl := lldars.NewDeliveryObject(serverId, localIP(conn), ServicePort, obj)
		conn.Write(sl.Marshal())
		log.Printf("Send Object > %s len: %d\n", conn.RemoteAddr().String(), sl.Length)
	}

	sl := lldars.NewEndOfDelivery(serverId, localIP(conn), ServicePort)
	conn.Write(sl.Marshal())
	return
}

func getObjectPaths(path string) []string {
	pat := path + "/*.zip"
	files, err := filepath.Glob(pat)
	Error(err)
	return files
}
