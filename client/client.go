package client

import (
	"context"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/silmin/lldars/pkg/lldars"
)

const (
	IntervalSeconds = 1
	TimeoutSeconds  = 10
	BroadcastAddr   = "192.168.100.255:60000"
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

	ip, _ := c.parseIpPort(conn.LocalAddr().String())
	log.Printf("conn.LocalAddr: %s", conn.LocalAddr().String())
	sl := lldars.NewGetObjectRequest(net.IP(ip).To4(), 0)
	msg := sl.Marshal()
	conn.Write(msg)

	buf := make([]byte, 1000)
	length, err := conn.Read(buf)
	Error(err)
	rl := lldars.Unmarshal(buf[:length])
	log.Printf("Recieve from: %v\tmsg: %s\tlen: %d\n", rl.Origin, rl.Payload, length)
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
	ip, _ := c.parseIpPort(conn.LocalAddr().String())
	udpAddr := &net.UDPAddr{
		IP:   net.ParseIP(ip),
		Port: 0,
	}
	udpLn, err := net.ListenUDP("udp", udpAddr)

	listenCtx, listenCancel := context.WithTimeout(ctx, time.Duration(TimeoutSeconds)*time.Second)
	defer listenCancel()
	go c.listenAck(listenCtx, udpLn, servicePort)

	addr, port := c.parseIpPort(udpLn.LocalAddr().String())
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

func (c *Client) parseIpPort(addr string) (string, string) {
	s := strings.Split(addr, ":")
	return s[0], s[1]
}

func Error(_err error) {
	if _err != nil {
		log.Panic(_err)
	}
}
