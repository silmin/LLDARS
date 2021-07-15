package client

import (
	"context"
	"log"
	"net"
	"os"
	"time"
)

const (
	IntervalSeconds = 1
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

func (c *Client) Broadcast(ctx context.Context, toAddr string, fromAddr string, listenAddr string) {
	fromEp, err := net.ResolveUDPAddr("udp", fromAddr)
	toEp, err := net.ResolveUDPAddr("udp", toAddr)
	conn, err := net.DialUDP("udp", fromEp, toEp)
	c.Error(err)
	defer conn.Close()
	log.Printf("Connected %s > %s\n", fromAddr, toAddr)

	// Get hostname
	msg, err := os.Hostname()
	c.Error(err)

	ticker := time.NewTicker(time.Duration(IntervalSeconds) * time.Second)
	defer ticker.Stop()

	listenCtx, listenCancel := context.WithCancel(ctx)
	go c.listenAck(listenCtx, listenCancel, listenAddr)
	for {
		select {
		case <-ctx.Done():
			log.Println("canceld broadcast")
			return
		case <-ticker.C:
			// Outbound message
			conn.Write([]byte(msg))
			log.Printf("Outbound %v > %v as “%s”\n", fromEp, toEp, msg)
		}
	}
}

func (c *Client) listenAck(ctx context.Context, cancel context.CancelFunc, addr string) {
	psock, err := net.Listen("tcp", addr)
	c.Error(err)
	log.Printf("Listened *:* > %s\n", addr)

	for {
		// Inbound message
		conn, err := psock.Accept()
		c.Error(err)
		go c.handleConnection(cancel, conn)
	}
}

func (c *Client) handleConnection(cancel context.CancelFunc, conn net.Conn) {
	defer conn.Close()
	buf := make([]byte, c.bufferSize)
	n, err := conn.Read(buf)
	c.Error(err)

	data := string(buf[:n])
	log.Printf("Receive msg: %s\n", data)
	cancel()
}

func (c *Client) Error(_err error) {
	if _err != nil {
		log.Panic(_err)
	}
}
