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
		log.Printf("Receive %v > %v\nmsg: %s\n", rl.Origin, listenAddr, rl.Payload)

		ipp := rl.Origin.String() + ":" + fmt.Sprintf("%d", rl.NextPort)
		ackAddr, err := net.ResolveUDPAddr("udp", ipp)
		sl := lldars.NewServerPortNotify(net.ParseIP(origin), 0)
		udpLn.WriteToUDP([]byte(sl.Marshal()), ackAddr)
	}
}

func Error(_err error) {
	if _err != nil {
		log.Panic(_err)
	}
}
