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
	SendObjectPath  = "./send_data/"
	SyncObjectPath  = "./sync_data/"
)

func Server(bcAddr string, origin string) {
	bcCtx, bcClose := context.WithCancel(context.Background())
	defer bcClose()

	go listenDiscoverBroadcast(bcCtx, bcAddr, origin)
	listenService()

	return
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
	defer conn.Close()
	buf := make([]byte, lldars.LLDARSLayerSize)
	l, err := conn.Read(buf)
	Error(err)
	msg := buf[:l]
	rl := lldars.Unmarshal(msg)
	log.Printf("Receive from: %v\tmsg: %s\n", rl.Origin, rl.Payload)

	if rl.Type == lldars.GetObjectRequest {
		buf := make([]byte, rl.Length)
		_, err := conn.Read(buf)
		Error(err)

		sendObjects(conn)
	}
	if rl.Type == lldars.SyncObjectRequest {
		buf := make([]byte, rl.Length)
		l, err := conn.Read(buf)
		Error(err)

		srvName := string(buf[:l])
		acceptSyncingObjects(conn, rl, srvName)
	}
	return
}

func sendObjects(conn net.Conn) {
	defer conn.Close()

	paths := getObjectPaths(SendObjectPath)
	for _, path := range paths {
		obj, err := ioutil.ReadFile(path)
		Error(err)
		sl := lldars.NewDeliveryObject(localIP(conn), ServicePort, obj)
		msg := sl.Marshal()
		conn.Write(msg)
		log.Printf("Send Object > %s len: %d\n", conn.RemoteAddr().String(), sl.Length)
	}

	sl := lldars.NewEndOfDelivery(localIP(conn), ServicePort)
	msg := sl.Marshal()
	conn.Write(msg)
	return
}

func acceptSyncingObjects(conn net.Conn, rl lldars.LLDARSLayer, srvName string) {
	defer conn.Close()

	// ack
	sl := lldars.NewAcceptSyncingObject(localIP(conn), ServicePort)
	msg := sl.Marshal()
	conn.Write(msg)

	path := SyncObjectPath + srvName + "/"
	objCnt := 0

	for {
		filename := path + fmt.Sprintf("%d.zip", objCnt)

		// header
		buf := make([]byte, lldars.LLDARSLayerSize)
		l, err := conn.Read(buf)
		Error(err)
		rl := lldars.Unmarshal(buf[:l])
		log.Printf("~ Recieve from: %v\tpayload-len: %d\n", rl.Origin, rl.Length)
		if rl.Type == lldars.EndOfDelivery {
			break
		} else if rl.Type != lldars.AcceptSyncingObject {
			continue
		}

		// object
		var obj []byte
		receivedBytes := 0
		for {
			bufSize := rl.Length - uint64(receivedBytes)
			if bufSize <= 0 {
				break
			}
			buf = make([]byte, bufSize)
			l, err = conn.Read(buf)
			Error(err)
			receivedBytes += l
			obj = append(obj, buf[:l]...)
			log.Printf("~ Read Parts %d (%d/%d)\n", l, len(obj), rl.Length)
		}

		if len(obj) != 0 {
			err = ioutil.WriteFile(filename, obj, 0644)
			Error(err)
			objCnt++
			log.Printf("Accept Object > %s, len: %d\n", filename, l)
		}
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

func localIP(conn net.Conn) net.IP {
	ipstr, _ := lldars.ParseIpPort(conn.LocalAddr().String())
	return net.ParseIP(ipstr).To4()
}

func Error(_err error) {
	if _err != nil {
		log.Panic(_err)
	}
}
