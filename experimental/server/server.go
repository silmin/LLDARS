package server

import (
	"log"
	"net"
)

const (
	BufferSize      = 1000
	IntervalSeconds = 1
)

func listenDelivering(addr string) error {
	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	Error(err)
	udpLn, err := net.ListenUDP("udp", udpAddr)
	Error(err)
	defer udpLn.Close()

	buf := make([]byte, BufferSize)
	for {
		length, from, err := udpLn.ReadFromUDP(buf)
		Error(err)
		msg := string(buf[:length])
		fromAddr := from.String()
		log.Printf("Inbound %v > %v as \"%s\"\n", fromAddr, addr, msg)
	}
}

func Error(_err error) {
	if _err != nil {
		log.Panic(_err)
	}
}

func Server(listenDeliveringAddr string) error {
	listenDelivering(listenDeliveringAddr)
	return nil
}
