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

func broadcastOutbound(cancelBroadcast <-chan struct{}) {
	from_udp_endpoint, err := net.ResolveUDPAddr("udp", OutboundFrom)
	to_udp_endpoint, err := net.ResolveUDPAddr("udp", OutboundTo)
	outbound_conn, err := net.DialUDP("udp", from_udp_endpoint, to_udp_endpoint)
	Error(err)
	defer outbound_conn.Close()
	fmt.Printf("Connected %s > %s\n", OutboundFrom, OutboundTo)

	// Get hostname
	msg, err := os.Hostname()
	Error(err)

	ticker := time.NewTicker(time.Duration(IntervalSeconds) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-cancelBroadcast:
			fmt.Println("canceld broadcast")
			return
		case <-ticker.C:
			// Outbound message
			outbound_conn.Write([]byte(msg))
			fmt.Printf("Outbound %v > %v as “%s”\n", from_udp_endpoint, to_udp_endpoint, msg)
		}
	}
}

func listenInbound(cancelBroadcast chan<- struct{}) {
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
		close(cancelBroadcast)
	}
}

func Error(_err error) {
	if _err != nil {
		panic(_err)
	}
}

func main() {
	cancelBroadcast := make(chan struct{})
	go broadcastOutbound(cancelBroadcast)
	listenInbound(cancelBroadcast)
}
