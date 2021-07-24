package server

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"path/filepath"

	"github.com/silmin/lldars/pkg/lldars"
)

const (
	IntervalSeconds = 1
	TimeoutSeconds  = 10
	ServicePort     = 60001
	ObjectPath      = "./data/"
)

func Server(listenAddr string, origin string) error {
	ldbCtx, ldbClose := context.WithCancel(context.Background())
	defer ldbClose()

	go listenDiscoverBroadcast(ldbCtx, listenAddr, origin)
	listenService()

	return nil
}

func listenService() {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", ServicePort))
	Error(err)
	for {
		conn, err := ln.Accept()
		Error(err)
		go handleService(conn)
	}
}

func handleService(conn net.Conn) {
	//buf := make([]byte, lldars.LLDARSLayerSize)
	buf := make([]byte, 1000)
	l, err := conn.Read(buf)
	Error(err)
	msg := buf[:l]
	rl := lldars.Unmarshal(msg)
	log.Printf("Receive from: %v\tmsg: %s\n", rl.Origin, rl.Payload)
	if rl.Type == lldars.GetObjectRequest {
		sendObjects(conn, rl)
	}
	return
}

func sendObjects(conn net.Conn, rl lldars.LLDARSLayer) {
	paths := getObjectPaths(ObjectPath)
	ip, _ := lldars.ParseIpPort(conn.LocalAddr().String())
	for _, path := range paths {
		obj, err := ioutil.ReadFile(path)
		Error(err)
		plen := uint64(len(obj))
		sl := lldars.NewDeliveryObject(net.ParseIP(ip).To4(), ServicePort, plen, obj)
		msg := sl.Marshal()
		conn.Write(msg)
		log.Printf("Send Object > %s : %s\n", conn.RemoteAddr().String(), rl.Payload)
	}
}

func getObjectPaths(path string) []string {
	pat := path + "*.zip"
	files, err := filepath.Glob(pat)
	Error(err)
	return files
}

func listenDiscoverBroadcast(ctx context.Context, listenAddr string, origin string) error {
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
		msg := buf[:length]
		rl := lldars.Unmarshal(msg)
		log.Printf("Receive from: %v\tmsg: %s\n", rl.Origin, rl.Payload)

		sl := lldars.NewServerPortNotify(net.ParseIP(origin), ServicePort)
		ipp := rl.Origin.String() + ":" + fmt.Sprintf("%d", rl.ServicePort)
		ackAddr, err := net.ResolveUDPAddr("udp", ipp)
		udpLn.WriteToUDP([]byte(sl.Marshal()), ackAddr)
		log.Printf("Ack to: %v\tmsg: %s\n", ackAddr.IP.String(), sl.Payload)
	}
}

func Error(_err error) {
	if _err != nil {
		log.Panic(_err)
	}
}
