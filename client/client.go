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
	nextAddrChan := make(chan string)
	go c.Broadcast(ctx, close, nextAddrChan)

	nextAddr := <-nextAddrChan
	log.Printf("service addr: %s\n", nextAddr)
}

func (c *Client) Broadcast(ctx context.Context, close context.CancelFunc, nextAddrChan chan<- string) {
	conn, err := net.Dial("udp", BroadcastAddr)
	Error(err)
	defer conn.Close()
	log.Printf("Connected > %s\n", BroadcastAddr)

	ticker := time.NewTicker(time.Duration(IntervalSeconds) * time.Second)
	defer ticker.Stop()

	// listen udp for ack
	ip, _ := c.parseAddr(conn.LocalAddr().String())
	udpAddr := &net.UDPAddr{
		IP:   net.ParseIP(ip),
		Port: 0,
	}
	udpLn, err := net.ListenUDP("udp", udpAddr)

	listenCtx, listenCancel := context.WithTimeout(ctx, time.Duration(TimeoutSeconds)*time.Second)
	go c.listenAck(listenCtx, udpLn, close, nextAddrChan)

	addr, port := c.parseAddr(udpLn.LocalAddr().String())
	Error(err)
	p, _ := strconv.Atoi(port)
	l := lldars.NewDiscoverBroadcast(net.ParseIP(addr).To4(), uint16(p))
	msg := l.Marshal()
	for {
		select {
		case <-ctx.Done():
			listenCancel()
			log.Println("End Broadcast")
			return
		case <-ticker.C:
			// broadcast
			conn.Write([]byte(msg))
			log.Printf("Cast > %v as “%s”\n", BroadcastAddr, l.Payload)
		}
	}
}

func (c *Client) listenAck(ctx context.Context, udpLn *net.UDPConn, close context.CancelFunc, nextAddrChan chan<- string) {
	log.Println("Start listenAck()")

	buf := make([]byte, c.bufferSize)
	length, err := udpLn.Read(buf)
	Error(err)
	msg := string(buf[:length])

	rl := lldars.Unmarshal([]byte(msg))
	log.Printf("Recieve from: %v\tmsg: %s\n", rl.Origin, rl.Payload)

	nextAddrChan <- rl.Origin.String() + ":" + fmt.Sprintf("%d", rl.ServicePort)
	close()
}

func (c *Client) parseAddr(addr string) (string, string) {
	s := strings.Split(addr, ":")
	return s[0], s[1]
}

func Error(_err error) {
	if _err != nil {
		log.Panic(_err)
	}
}
