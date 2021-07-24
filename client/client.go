package client

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"strconv"
	"time"

	"github.com/silmin/lldars/pkg/lldars"
)

const (
	IntervalSeconds   = 1
	TimeoutSeconds    = 10
	BroadcastAddr     = "192.168.100.255:60000"
	ReceiveObjectPath = "../receive_data/"
)

type Client struct {
	fromAddr   string
	toAddr     string
	bufferSize int
}

func NewClient(buf int) *Client {
	return &Client{
		bufferSize: buf,
	}
}

func (c *Client) DoAct() {
	ctx, close := context.WithTimeout(context.Background(), time.Duration(TimeoutSeconds)*time.Second)
	defer close()
	serviceAddr := make(chan string)
	go c.Broadcast(ctx, serviceAddr)

	addr := <-serviceAddr
	log.Printf("service addr: %s\n", addr)
	close() // close broadcast

	c.getObjects(addr)

	return
}

func (c *Client) getObjects(addr string) {
	conn, err := net.Dial("tcp", addr)
	Error(err)
	defer conn.Close()

	ip, _ := lldars.ParseIpPort(conn.LocalAddr().String())
	log.Printf("conn.LocalAddr: %s", conn.LocalAddr().String())
	sl := lldars.NewGetObjectRequest(net.ParseIP(ip).To4(), 0)
	msg := sl.Marshal()
	conn.Write(msg)

	fc := 0

	for {
		filename := ReceiveObjectPath + fmt.Sprintf("%d.zip", fc)
		buf := make([]byte, lldars.LLDARSLayerSize)
		length, err := conn.Read(buf)
		Error(err)
		rl := lldars.Unmarshal(buf[:length])
		log.Printf("Recieve from: %v\tlength: %d\n", rl.Origin, rl.Length)
		if rl.Type == lldars.EndDeliveryObject {
			break
		}
		buf = make([]byte, rl.Length)
		length, err = conn.Read(buf)
		Error(err)

		obj := buf[:length]
		err = ioutil.WriteFile(filename, obj, 0644)
		Error(err)
		fc++
		log.Printf("Read Object > %s", filename)
	}

	log.Println("End getObjects()")
	return
}

func (c *Client) Broadcast(ctx context.Context, servicePort chan<- string) {
	conn, err := net.Dial("udp", BroadcastAddr)
	Error(err)
	defer conn.Close()
	log.Printf("Connected > %s\n", BroadcastAddr)

	ticker := time.NewTicker(time.Duration(IntervalSeconds) * time.Second)
	defer ticker.Stop()

	// listen udp for ack
	ip, _ := lldars.ParseIpPort(conn.LocalAddr().String())
	udpAddr := &net.UDPAddr{
		IP:   net.ParseIP(ip),
		Port: 0,
	}
	udpLn, err := net.ListenUDP("udp", udpAddr)

	listenCtx, listenCancel := context.WithTimeout(ctx, time.Duration(TimeoutSeconds)*time.Second)
	defer listenCancel()
	go c.listenAck(listenCtx, udpLn, servicePort)

	addr, port := lldars.ParseIpPort(udpLn.LocalAddr().String())
	Error(err)
	p, _ := strconv.Atoi(port)
	l := lldars.NewDiscoverBroadcast(net.ParseIP(addr).To4(), uint16(p))
	msg := l.Marshal()
	for {
		select {
		case <-ctx.Done():
			log.Println("End Broadcast")
			return
		case <-ticker.C:
			// broadcast
			conn.Write(msg)
			log.Printf("Cast > %v : “%s”\n", BroadcastAddr, l.Payload)
		}
	}
}

func (c *Client) listenAck(ctx context.Context, udpLn *net.UDPConn, nextAddrChan chan<- string) {
	log.Println("Start listenAck()")

	buf := make([]byte, c.bufferSize)
	length, err := udpLn.Read(buf)
	Error(err)
	msg := string(buf[:length])

	rl := lldars.Unmarshal([]byte(msg))
	log.Printf("Recieve from: %v\tmsg: %s\n", rl.Origin, rl.Payload)

	nextAddrChan <- rl.Origin.String() + ":" + fmt.Sprintf("%d", rl.ServicePort)
}

func Error(_err error) {
	if _err != nil {
		log.Panic(_err)
	}
}
