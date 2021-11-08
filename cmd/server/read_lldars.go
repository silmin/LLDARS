package main

import (
	"net"

	"github.com/silmin/lldars/pkg/lldars"
)

func ReadLLDARSHeader(conn net.Conn) lldars.LLDARSLayer {
	return lldars.Unmarshal(readConnData(conn, lldars.LLDARSLayerSize))
}

func ReadLLDARSPayload(conn net.Conn, length uint64) []byte {
	return readConnData(conn, int(length))
}

func readConnData(conn net.Conn, length int) []byte {
	var data []byte
	for {
		bufSize := int(length) - len(data)
		if bufSize <= 0 {
			break
		}
		buf := make([]byte, bufSize)
		l, err := conn.Read(buf)
		data = append(data, buf[:l]...)
		Error(err)
	}
	return data
}
