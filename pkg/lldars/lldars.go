package lldars

import (
	"encoding/binary"
	"net"
	"strings"
)

type LLDARSLayer struct {
	Type        LLDARSLayerType
	Origin      net.IP
	ServicePort uint16
	Length      uint64
	Payload     []byte
}

type LLDARSLayerType uint8

const (
	DiscoverBroadcast LLDARSLayerType = iota
	ServicePortNotify
	GetObjectRequest
	DeliveryObject
)

const (
	LLDARSLayerSize          = 1 + 4 + 2 + 16
	DiscoverBroadcastPayload = "Is available LLDARS server on this network ?"
	ServicePortNotifyPayload = "--NotifyServerPortPayload--"
	GetObjectRequestPayload  = "--GetObjectRequestPayload--"
	DeliveryObjectPayload    = "--DeliveryObjectPayload--"
)

func NewDiscoverBroadcast(origin net.IP, sp uint16) LLDARSLayer {
	length := uint64(len(DiscoverBroadcastPayload))
	return NewLLDARSPacket(origin, sp, length, DiscoverBroadcast, DiscoverBroadcastPayload)
}

func NewServerPortNotify(origin net.IP, sp uint16) LLDARSLayer {
	length := uint64(len(DiscoverBroadcastPayload))
	return NewLLDARSPacket(origin, sp, length, ServicePortNotify, ServicePortNotifyPayload)
}

func NewGetObjectRequest(origin net.IP, sp uint16) LLDARSLayer {
	length := uint64(len(DiscoverBroadcastPayload))
	return NewLLDARSPacket(origin, sp, length, GetObjectRequest, GetObjectRequestPayload)
}

func NewDeliveryObject(origin net.IP, sp uint16, l uint64) LLDARSLayer {
	return NewLLDARSPacket(origin, sp, l, DeliveryObject, DeliveryObjectPayload)
}

func NewLLDARSPacket(origin net.IP, sp uint16, l uint64, t LLDARSLayerType, p string) LLDARSLayer {
	return LLDARSLayer{
		Type:        t,
		Origin:      origin,
		ServicePort: sp,
		Length:      l,
		Payload:     []byte(p),
	}
}

func (l *LLDARSLayer) Marshal() []byte {
	buf := make([]byte, LLDARSLayerSize)
	buf[0] = byte(l.Type)
	binary.BigEndian.PutUint32(buf[1:], ip2int(l.Origin))
	binary.BigEndian.PutUint16(buf[5:], l.ServicePort)
	binary.BigEndian.PutUint64(buf[7:], uint64(l.Length))
	buf = append(buf, l.Payload...)
	return buf
}

func Unmarshal(buf []byte) LLDARSLayer {
	var l LLDARSLayer
	l.Type = LLDARSLayerType(buf[0])
	l.Origin = int2ip(binary.BigEndian.Uint32(buf[1:]))
	l.ServicePort = binary.BigEndian.Uint16(buf[5:])
	l.Length = binary.BigEndian.Uint64(buf[7:])
	l.Payload = buf[15:]
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

func ParseIpPort(addr string) (string, string) {
	s := strings.Split(addr, ":")
	return s[0], s[1]
}
