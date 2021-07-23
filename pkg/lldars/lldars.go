package lldars

import (
	"encoding/binary"
	"net"
)

type LLDARSLayer struct {
	Type        LLDARSLayerType
	Origin      net.IP
	ServicePort uint16
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
	LLDARSLayerSize          = 1 + 4 + 2
	DiscoverBroadcastPayload = "Is available LLDARS server on this network ?"
	ServicePortNotifyPayload = "--NotifyServerPortPayload--"
	GetObjectRequestPayload  = "--GetObjectRequestPayload--"
	DeliveryObjectPayload    = "--DeliveryObjectPayload--"
)

func NewDiscoverBroadcast(origin net.IP, sp uint16) LLDARSLayer {
	return NewLLDARSPacket(origin, sp, DiscoverBroadcast, DiscoverBroadcastPayload)
}

func NewServerPortNotify(origin net.IP, sp uint16) LLDARSLayer {
	return NewLLDARSPacket(origin, sp, ServicePortNotify, ServicePortNotifyPayload)
}

func NewGetObjectRequest(origin net.IP, sp uint16) LLDARSLayer {
	return NewLLDARSPacket(origin, sp, GetObjectRequest, GetObjectRequestPayload)
}

func NewDeliveryObject(origin net.IP, sp uint16) LLDARSLayer {
	return NewLLDARSPacket(origin, sp, DeliveryObject, DeliveryObjectPayload)
}

func NewLLDARSPacket(origin net.IP, sp uint16, t LLDARSLayerType, p string) LLDARSLayer {
	return LLDARSLayer{
		Type:        t,
		Origin:      origin,
		ServicePort: sp,
		Payload:     []byte(p),
	}
}

func (l *LLDARSLayer) Marshal() []byte {
	buf := make([]byte, LLDARSLayerSize)
	buf[0] = byte(l.Type)
	binary.BigEndian.PutUint32(buf[1:], ip2int(l.Origin))
	binary.BigEndian.PutUint16(buf[5:], l.ServicePort)
	buf = append(buf, l.Payload...)
	return buf
}

func Unmarshal(buf []byte) LLDARSLayer {
	var l LLDARSLayer
	l.Type = LLDARSLayerType(buf[0])
	l.Origin = int2ip(binary.BigEndian.Uint32(buf[1:]))
	l.ServicePort = binary.BigEndian.Uint16(buf[5:])
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
