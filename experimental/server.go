package main

import (
	"fmt"
	"net"
)

const (
	InboundTo    = "192.168.100.255:60000"
	OutboundFrom = "192.168.100.1:60000"
	OutboundTo   = "192.168.100.255:60000"
	BufferByte   = 64
)

func sendAck(from string) {

}

func listenBroadcast(cancelBroadcast chan<- struct{}) {
	to_udp_endpoint, err := net.ResolveUDPAddr("udp", InboundTo)
	Error(err)

	inbound, err := net.ListenUDP("udp", to_udp_endpoint)
	Error(err)
	defer inbound.Close()
	fmt.Printf("Listened *:* > %s\n", InboundTo)

	buffer := make([]byte, BufferByte)
	for {
		// Inbound message
		length, addr, err := inbound.ReadFrom(buffer)
		Error(err)
		msg := string(buffer[:length])

		inbound_from := addr.(*net.UDPAddr).String()
		if inbound_from == OutboundFrom {
			continue
		}

		fmt.Printf("Inbound %v > %v as “%s”\n", inbound_from, InboundTo, msg)

		go sendAck(inbound_from)
	}
}

func Error(_err error) {
	if _err != nil {
		panic(_err)
	}
}

func main() {
	cancelBroadcast := make(chan struct{})
	listenBroadcast(cancelBroadcast)
}
