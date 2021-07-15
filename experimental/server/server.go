package server

import (
	"log"
	"net"
)

const (
	BufferSize      = 1000
	IntervalSeconds = 1
)

func Server(listenDeliveringAddr string) error {
	listenDeliveringRequest(listenDeliveringAddr)
	return nil
}

func listenDeliveringRequest(addr string) error {
	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	Error(err)
	udpLn, err := net.ListenUDP("udp", udpAddr)
	Error(err)
	defer udpLn.Close()

	log.Printf("Listened Delivering Request *:* udp > %s\n", addr)
	buf := make([]byte, BufferSize)
	for {
		length, from, err := udpLn.ReadFrom(buf)
		Error(err)
		msg := string(buf[:length])
		fromAddr := from.(*net.UDPAddr).String()
		log.Printf("Inbound %v > %v as \"%s\"\n", fromAddr, addr, msg)
	}
}

func Error(_err error) {
	if _err != nil {
		log.Panic(_err)
	}
}
