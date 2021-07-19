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

func Server(listenAddr string, origin string) error {
	listenDiscoverBroadcast(listenAddr, origin)
	return nil
}

func listenDiscoverBroadcast(listenAddr string, origin string) error {
	udpAddr, err := net.ResolveUDPAddr("udp", listenAddr)
	Error(err)
	udpLn, err := net.ListenUDP("udp", udpAddr)
	Error(err)
	defer udpLn.Close()

	log.Printf("Listened Delivering Requests *:* udp > %s\n", listenAddr)
	buf := make([]byte, lldars.LLDARSLayerSize)
	for {
		length, err := udpLn.Read(buf)
		Error(err)
		msg := string(buf[:length])
		rl := lldars.Unmarshal([]byte(msg))
		log.Printf("Receive from: %v\tmsg: %s\n", rl.Origin, rl.Payload)

		sl := lldars.NewServerPortNotify(net.ParseIP(origin), 0)
		ipp := rl.Origin.String() + ":" + fmt.Sprintf("%d", rl.NextPort)
		ackAddr, err := net.ResolveUDPAddr("udp", ipp)
		udpLn.WriteToUDP([]byte(sl.Marshal()), ackAddr)
		log.Printf("Ack to: %v\tmsg: %s\n", ackAddr.IP.String(), rl.Payload)
	}
}

func Error(_err error) {
	if _err != nil {
		log.Panic(_err)
	}
}
