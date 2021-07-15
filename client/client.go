package client

import (
	"context"
	"log"
	"net"
	"strings"
	"time"
)

const (
	IntervalSeconds = 1
	TimeoutSeconds  = 10
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

func (c *Client) Broadcast(ctx context.Context, close context.CancelFunc, toAddr string) {
	conn, err := net.Dial("udp", toAddr)
	Error(err)
	defer conn.Close()
	log.Printf("Connected > %s\n", toAddr)

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
	go c.listenAck(listenCtx, udpLn, close)

	msg := udpLn.LocalAddr().String()
	for {
		select {
		case <-ctx.Done():
			listenCancel()
			log.Println("End Broadcast")
			return
		case <-ticker.C:
			// broadcast
			conn.Write([]byte(msg))
			log.Printf("Cast > %v as “%s”\n", toAddr, msg)
		}
	}
}

func (c *Client) listenAck(ctx context.Context, udpLn *net.UDPConn, close context.CancelFunc) {
	log.Println("Start listenAck()")

	buf := make([]byte, c.bufferSize)
	length, from, err := udpLn.ReadFrom(buf)
	Error(err)
	msg := string(buf[:length])
	fromAddr := from.(*net.UDPAddr).String()

	log.Printf("Recieve from: %v\tmsg: %s\n", fromAddr, msg)
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
