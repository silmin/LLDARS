package server

import (
	"fmt"
	"log"
	"net"

	"github.com/silmin/lldars/pkg/lldars"
)

const (
	IntervalSeconds = 1
	TimeoutSeconds  = 10
)

func Server(listenDeliveringAddr string) error {
	listenDeliveryRequest(listenDeliveringAddr)
	return nil
}

func listenDeliveryRequest(addr string) error {
	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	Error(err)
	udpLn, err := net.ListenUDP("udp", udpAddr)
	Error(err)
	defer udpLn.Close()

	log.Printf("Listened Delivering Requests *:* udp > %s\n", addr)
	buf := make([]byte, lldars.LLDARSLayerSize)
	for {
		length, err := udpLn.Read(buf)
		Error(err)
		msg := string(buf[:length])
		rl := lldars.Unmarshal([]byte(msg))
		log.Printf("Receive %v > %v\nmsg: %s\n", rl.Origin, addr, rl.Payload)

		ip := rl.Origin.String() + ":" + fmt.Sprintf("%d", rl.NextPort)
		ackAddr, err := net.ResolveUDPAddr("udp", ip)
		sl := lldars.NewServerPortNotify(net.ParseIP(addr), 0)
		udpLn.WriteToUDP([]byte(sl.Marshal()), ackAddr)
	}
}

func Error(_err error) {
	if _err != nil {
		log.Panic(_err)
	}
}
