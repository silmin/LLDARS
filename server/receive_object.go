package server

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"time"

	"github.com/silmin/lldars/pkg/lldars"
)

func receiveObjects(conn net.Conn, serverId uint32, path string) {
	log.Println("--Start receiving objects--")
	for {
		filename := path + genFilename()

		// header
		buf := make([]byte, lldars.LLDARSLayerSize)
		l, err := conn.Read(buf)
		Error(err)
		rl := lldars.Unmarshal(buf[:l])
		log.Printf("Recieve from: %v\tpayload-len: %d\n", rl.Origin, rl.Length)
		log.Printf("from serverId: %d\n", rl.ServerId)
		if rl.Type == lldars.EndOfDelivery {
			sl := lldars.NewReceivedObjects(serverId, localIP(conn), ServicePort)
			conn.Write(sl.Marshal())
			log.Println("--End receiving objects--")
			return
		} else if rl.Type != lldars.DeliveryObject {
			continue
		}

		// object
		var obj []byte
		receivedBytes := 0
		for {
			bufSize := rl.Length - uint64(receivedBytes)
			if bufSize <= 0 {
				break
			}
			buf = make([]byte, bufSize)
			l, err = conn.Read(buf)
			Error(err)
			receivedBytes += l
			obj = append(obj, buf[:l]...)
			log.Printf("Read Parts %d (%d/%d)\n", l, len(obj), rl.Length)
		}

		if len(obj) != 0 {
			err = ioutil.WriteFile(filename, obj, 0644)
			Error(err)
			log.Printf("Receive Object > %s, len: %d\n", filename, rl.Length)
		}
	}
}

func genFilename() string {
	t := time.Now()
	return fmt.Sprintf("%s.zip", t.Format("20060102T150405.000"))
}
