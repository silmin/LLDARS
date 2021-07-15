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

func (c *Client) Broadcast(ctx context.Context, toAddr string, fromAddr string, listenAddr string) {
	fromEp, err := net.ResolveUDPAddr("udp", fromAddr)
	toEp, err := net.ResolveUDPAddr("udp", toAddr)
	conn, err := net.DialUDP("udp", fromEp, toEp)
	Error(err)
	defer conn.Close()
	log.Printf("Connected %s > %s\n", fromAddr, toAddr)

	// Get hostname
	msg, err := os.Hostname()
	Error(err)

	ticker := time.NewTicker(time.Duration(IntervalSeconds) * time.Second)
	defer ticker.Stop()

	listenCtx, listenCancel := context.WithTimeout(ctx, time.Duration(TimeoutSeconds)*time.Second)
	go c.listenAck(listenCtx, conn)
	for {
		select {
		case <-ctx.Done():
			listenCancel()
			log.Println("broadcast timeout")
			return
		case <-ticker.C:
			// Outbound message
			conn.Write([]byte(msg))
			log.Printf("Cast %v > %v as “%s”\n", fromEp, toEp, msg)
		}
	}
}

func (c *Client) listenAck(ctx context.Context, conn *net.UDPConn) {
	buf := make([]byte, c.bufferSize)

	length, from, err := conn.ReadFrom(buf)
	Error(err)
	msg := string(buf[:length])
	fromAddr := from.(*net.UDPAddr).String()

	log.Printf("Recieve %v > %v\nmsg: %s\n", fromAddr, conn.LocalAddr().String(), msg)
}

func Error(_err error) {
	if _err != nil {
		log.Panic(_err)
	}
}
