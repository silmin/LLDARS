package lldars

import (
	"encoding/binary"
	"net"
)

type LLDARSLayer struct {
	Type     LLDARSLayerType
	Origin   net.IP
	NextPort uint16
	Payload  []byte
}

type LLDARSLayerType uint8

const (
	DiscoverBroadcast LLDARSLayerType = iota
	ServerPortNotify
)

const (
	LLDARSLayerSize          = 1 + 4 + 2
	DiscoverBroadcastPayload = "Is available LLDARS server on this network ?"
	ServerPortNotifyPayload  = "--NotifyServerPortPayload--"
)

func NewDiscoverBroadcast(origin net.IP, next uint16) LLDARSLayer {
	return LLDARSLayer{
		Type:     DiscoverBroadcast,
		Origin:   origin,
		NextPort: next,
		Payload:  []byte(DiscoverBroadcastPayload),
	}
}

func NewServerPortNotify(origin net.IP, next uint16) LLDARSLayer {
	return LLDARSLayer{
		Type:     ServerPortNotify,
		Origin:   origin,
		NextPort: next,
		Payload:  []byte(ServerPortNotifyPayload),
	}
}

func (l *LLDARSLayer) Marshal() []byte {
	buf := make([]byte, LLDARSLayerSize)
	buf[0] = byte(l.Type)
	binary.BigEndian.PutUint32(buf[1:], ip2int(l.Origin))
	binary.BigEndian.PutUint16(buf[5:], l.NextPort)
	buf = append(buf, l.Payload...)
	return buf
}

func Unmarshal(buf []byte) LLDARSLayer {
	var l LLDARSLayer
	l.Type = LLDARSLayerType(buf[0])
	l.Origin = int2ip(binary.BigEndian.Uint32(buf[1:]))
	l.NextPort = binary.BigEndian.Uint16(buf[5:])
	l.Payload = buf[7:]
	return l
}

func ip2int(ip net.IP) uint32 {
	if len(ip) == 16 {
		return binary.BigEndian.Uint32(ip[12:16])
	}
	return binary.BigEndian.Uint32(ip)
}

func int2ip(n uint32) net.IP {
	ip := make(net.IP, 4)
	binary.BigEndian.PutUint32(ip, n)
	return ip
}
