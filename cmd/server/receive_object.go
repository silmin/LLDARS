package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"time"

	"github.com/silmin/lldars/pkg/lldars"
)

func receiveObjects(conn net.Conn, path string) {
	log.Println("--Start receiving objects--")
	for {
		// header
		rl := ReadLLDARSHeader(conn)
		log.Printf("Recieve from: %v\tpayload-len: %d\n", rl.Origin, rl.Length)
		log.Printf("from serverId: %d\n", rl.ServerId)
		if rl.Type == lldars.EndOfDelivery {
			log.Println("--End receiving objects--")
			return
		} else if rl.Type != lldars.DeliveryObject {
			continue
		}

		filename := path + genFilename()

		// object
		obj := ReadLLDARSPayload(conn, rl.Length)
		if len(obj) != 0 {
			err := ioutil.WriteFile(filename, obj, 0644)
			Error(err)
			log.Printf("Receive Object > %s, len: %d\n", filename, rl.Length)
		}
	}
}

func genFilename() string {
	t := time.Now()
	return fmt.Sprintf("%s.zip", t.Format("20060102T150405.000"))
}
