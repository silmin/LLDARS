package server

import (
	"net"

	"github.com/silmin/lldars/pkg/lldars"
)

func readLLDARSHeader(conn net.Conn) lldars.LLDARSLayer {
	var header []byte
	for {
		bufSize := lldars.LLDARSLayerSize - len(header)
		if bufSize <= 0 {
			break
		}
		buf := make([]byte, bufSize)
		l, err := conn.Read(buf)
		header = append(header, buf[:l]...)
		Error(err)
	}
	rl := lldars.Unmarshal(header)

	return rl
}

func readLLDARSPayload(conn net.Conn, length uint64) []byte {
	var payload []byte
	for {
		bufSize := int(length) - len(payload)
		if bufSize <= 0 {
			break
		}
		buf := make([]byte, bufSize)
		l, err := conn.Read(buf)
		payload = append(payload, buf[:l]...)
		Error(err)
	}
	return payload
}
