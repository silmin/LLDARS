package main

import (
	"fmt"
	"net"
	"os"
	"time"
)

var InboundTo string = "192.168.100.255:60000"
var OutboundFrom string = "192.168.100.1:60000"
var OutboundTo string = "192.168.100.255:60000"
var BufferByte int = 64
var IntervalSeconds int = 1 // * 1 sec

func listenInbound() {
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

		//inbound_from := addr.(*net.UDPAddr).String()
		inbound_from := addr.String()

		fmt.Printf("Inbound %v > %v as “%s”\n", inbound_from, InboundTo, msg)
	}
}

func broadcastOutbound() {
	from_udp_endpoint, err := net.ResolveUDPAddr("udp", OutboundFrom)
	to_udp_endpoint, err := net.ResolveUDPAddr("udp", OutboundTo)
	outbound, err := net.DialUDP("udp", from_udp_endpoint, to_udp_endpoint)
	Error(err)
	defer outbound.Close()
	fmt.Printf("Connected %s > %s\n", OutboundFrom, OutboundTo)

	// Get hostname
	msg, err := os.Hostname()
	Error(err)

	for {
		time.Sleep(time.Duration(IntervalSeconds) * time.Second)

		// Outbound message
		outbound.Write([]byte(msg))
		fmt.Printf("Outbound %v > %v as “%s”\n", from_udp_endpoint, to_udp_endpoint, msg)
	}
}

func Error(_err error) {
	if _err != nil {
		panic(_err)
	}
}

func main() {
	go listenInbound()
	broadcastOutbound()
}
