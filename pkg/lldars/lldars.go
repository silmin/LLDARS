package lldars

import (
	"encoding/binary"
	"net"
	"strings"
)

type LLDARSLayer struct {
	Type        LLDARSLayerType
	ServerId    uint32
	Origin      net.IP
	ServicePort uint16
	Length      uint64
	Payload     []byte
}

type LLDARSLayerType uint8
type LLDARSServeMode uint8

const (
	ServicePort                = 60000
	ClientBCPort               = 60001
	ServerBCPort               = 60002
	LLDARSLayerSize            = 1 + 4 + net.IPv4len + 2 + 8
	DiscoverBroadcastPayload   = "Is available LLDARS server on this network ?"
	ServicePortNotifyPayload   = "--NotifyServerPortPayload--"
	GetObjectRequestPayload    = "--GetObjectRequestPayload--"
	DeliveryObjectPayload      = "--DeliveryObjectPayload--"
	EndOfDeliveryPayload       = "--EndOfDelivery--"
	ReceivedObjectsPayload     = "--ReceivedObjectsPayload--"
	BackupObjectRequestPayload = "--BackupObjectRequestPayload--"
	AcceptBackupObjectPayload  = "--AcceptBackupObjectPayload--"
	RejectBackupObjectPayload  = "--RejectBackupObjectPayload--"
	SyncObjectRequestPayload   = "--SyncObjectRequestPayload--"
	AcceptSyncObjectPayload    = "--AcceptSyncObjectPayload--"
	RejectSyncObjectPayload    = "--RejectSyncObjectPayload--"
)

func NewDiscoverBroadcast(id uint32, origin net.IP, sp uint16) LLDARSLayer {
	l := uint64(len(DiscoverBroadcastPayload))

	return NewLLDARSPacket(id, origin, sp, l, DiscoverBroadcast, []byte(DiscoverBroadcastPayload))
}

func NewServerPortNotify(id uint32, origin net.IP, sp uint16) LLDARSLayer {
	l := uint64(len(ServicePortNotifyPayload))
	return NewLLDARSPacket(id, origin, sp, l, ServicePortNotify, []byte(ServicePortNotifyPayload))
}

func NewGetObjectRequest(id uint32, origin net.IP, sp uint16) LLDARSLayer {
	l := uint64(len(GetObjectRequestPayload))
	return NewLLDARSPacket(id, origin, sp, l, GetObjectRequest, []byte(GetObjectRequestPayload))
}

func NewDeliveryObject(id uint32, origin net.IP, sp uint16, obj []byte) LLDARSLayer {
	l := uint64(len(obj))
	return NewLLDARSPacket(id, origin, sp, l, DeliveryObject, obj)
}

func NewEndOfDelivery(id uint32, origin net.IP, sp uint16) LLDARSLayer {
	l := uint64(len(EndOfDeliveryPayload))
	return NewLLDARSPacket(id, origin, sp, l, EndOfDelivery, []byte(EndOfDeliveryPayload))
}

func NewReceivedObjects(id uint32, origin net.IP, sp uint16) LLDARSLayer {
	l := uint64(len(ReceivedObjectsPayload))
	return NewLLDARSPacket(id, origin, sp, l, ReceivedObjects, []byte(ReceivedObjectsPayload))
}

func NewBackupObjectRequest(id uint32, origin net.IP, sp uint16) LLDARSLayer {
	l := uint64(len(BackupObjectRequestPayload))
	return NewLLDARSPacket(id, origin, sp, l, BackupObjectRequest, []byte(BackupObjectRequestPayload))
}

func NewAcceptBackupObject(id uint32, origin net.IP, sp uint16) LLDARSLayer {
	l := uint64(len(AcceptBackupObjectPayload))
	return NewLLDARSPacket(id, origin, sp, l, AcceptBackupObject, []byte(AcceptBackupObjectPayload))
}

func NewRejectBackupObject(id uint32, origin net.IP, sp uint16) LLDARSLayer {
	l := uint64(len(RejectBackupObjectPayload))
	return NewLLDARSPacket(id, origin, sp, l, RejectBackupObject, []byte(RejectBackupObjectPayload))
}

func NewSyncObjectRequest(id uint32, origin net.IP, sp uint16) LLDARSLayer {
	l := uint64(len(SyncObjectRequestPayload))
	return NewLLDARSPacket(id, origin, sp, l, SyncObjectRequest, []byte(SyncObjectRequestPayload))
}

func NewAcceptSyncObject(id uint32, origin net.IP, sp uint16) LLDARSLayer {
	l := uint64(len(AcceptSyncObjectPayload))
	return NewLLDARSPacket(id, origin, sp, l, AcceptSyncObject, []byte(AcceptSyncObjectPayload))
}

func NewRejectSyncObject(id uint32, origin net.IP, sp uint16) LLDARSLayer {
	l := uint64(len(RejectSyncObjectPayload))
	return NewLLDARSPacket(id, origin, sp, l, RejectSyncObject, []byte(RejectSyncObjectPayload))
}

func NewLLDARSPacket(id uint32, origin net.IP, sp uint16, l uint64, t LLDARSLayerType, p []byte) LLDARSLayer {
	return LLDARSLayer{
		Type:        t,
		ServerId:    id,
		Origin:      origin,
		ServicePort: sp,
		Length:      l,
		Payload:     []byte(p),
	}
}

func (l *LLDARSLayer) Marshal() []byte {
	buf := make([]byte, LLDARSLayerSize)
	buf[0] = byte(l.Type)
	binary.BigEndian.PutUint32(buf[1:], l.ServerId)
	binary.BigEndian.PutUint32(buf[5:], ip2int(l.Origin))
	binary.BigEndian.PutUint16(buf[9:], l.ServicePort)
	binary.BigEndian.PutUint64(buf[11:], uint64(l.Length))
	buf = append(buf, l.Payload...)
	return buf
}

func Unmarshal(buf []byte) LLDARSLayer {
	var l LLDARSLayer
	l.Type = LLDARSLayerType(buf[0])
	l.ServerId = binary.BigEndian.Uint32(buf[1:])
	l.Origin = int2ip(binary.BigEndian.Uint32(buf[5:]))
	l.ServicePort = binary.BigEndian.Uint16(buf[9:])
	l.Length = binary.BigEndian.Uint64(buf[11:])
	l.Payload = buf[19:]
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
	if len(s) < 2 {
		return s[0], ""
	} else {
		return s[0], s[1]
	}
}

func IsEqualIP(a net.IP, b net.IP) bool {
	return ip2int(a) == ip2int(b)
}
