package server

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"

	"github.com/silmin/lldars/pkg/lldars"
)

func receiveSyncObjects(conn net.Conn, rl lldars.LLDARSLayer, serverId uint32) {
	defer conn.Close()

	sl := lldars.NewAcceptSyncingObject(serverId, localIP(conn), ServicePort)
	msg := sl.Marshal()
	conn.Write(msg)

	path := fmt.Sprintf("%s/%d/", SyncObjectPath, serverId)
	objCnt := 0

	for {
		filename := path + fmt.Sprintf("%d.zip", objCnt)

		// header
		buf := make([]byte, lldars.LLDARSLayerSize)
		l, err := conn.Read(buf)
		Error(err)
		rl := lldars.Unmarshal(buf[:l])
		log.Printf("~ Recieve from: %v\tpayload-len: %d\n", rl.Origin, rl.Length)
		if rl.Type == lldars.EndOfDelivery {
			break
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
			log.Printf("~ Read Parts %d (%d/%d)\n", l, len(obj), rl.Length)
		}

		if len(obj) != 0 {
			err = ioutil.WriteFile(filename, obj, 0644)
			Error(err)
			objCnt++
			log.Printf("Accept Object > %s, len: %d\n", filename, l)
		}
	}
}
