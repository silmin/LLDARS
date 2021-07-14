package main

import (
	"fmt"
	"net"
	"os"
	"time"
)

const (
	InboundTo       = "192.168.100.255:60000"
	OutboundFrom    = "192.168.100.1:60000"
	OutboundTo      = "192.168.100.255:60000"
	BufferByte      = 1024
	IntervalSeconds = 1
)

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

func handleConnection(conn net.Conn, cancelBroadcast chan<- struct{}) {
	defer conn.Close()
	buf := make([]byte, BufferByte)
	n, err := conn.Read(buf)
	if err != nil {
		fmt.Printf("Read Error: %s\n", err)
	}
	data := string(buf[:n])
	fmt.Printf("Receive msg: %s\n", data)
	close(cancelBroadcast)
}

func listenInbound(cancelBroadcast chan<- struct{}) {
	psock, err := net.Listen("tcp", ":60000")
	Error(err)
	fmt.Printf("Listened *:* > %s\n", InboundTo)

	for {
		// Inbound message
		conn, err := psock.Accept()
		Error(err)
		go handleConnection(conn, cancelBroadcast)
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
