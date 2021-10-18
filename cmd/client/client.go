package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"time"

	"github.com/silmin/lldars/pkg/lldars"
)

const (
	IntervalSeconds   = 1
	TimeoutSeconds    = 10
	BroadcastAddr     = "192.168.100.255:60000"
	ReceiveObjectPath = "./receive_data/"
)

type Client struct{}

func NewClient() *Client {
	return &Client{}
}

func (c *Client) DoAct() {
	ctx, close := context.WithTimeout(context.Background(), time.Duration(TimeoutSeconds)*time.Second)
	defer close()

	serviceAddr := make(chan string)
	go DiscoverBroadcast(ctx, 0, serviceAddr)

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
	sl := lldars.NewGetObjectRequest(0, net.ParseIP(ip).To4(), 0)
	msg := sl.Marshal()
	conn.Write(msg)

	// receive objects
	for {
		filename := ReceiveObjectPath + genFilename()

		// header
		rl := ReadLLDARSHeader(conn)
		log.Printf("Recieve from: %v\tpayload-len: %d\n", rl.Origin, rl.Length)
		log.Printf("serverId: %d\n", rl.ServerId)
		if rl.Type == lldars.EndOfDelivery {
			sl := lldars.NewReceivedObjects(0, localConnIP(conn), 0)
			conn.Write(sl.Marshal())
			log.Println("--End receiving objects--")
			break
		} else if rl.Type != lldars.DeliveryObject {
			continue
		}

		// object
		obj := ReadLLDARSPayload(conn, rl.Length)
		if len(obj) != 0 {
			err = ioutil.WriteFile(filename, obj, 0644)
			Error(err)
			log.Printf("Read Object > %s, len: %d\n", filename, rl.Length)
		}
	}

	log.Println("End getObjects()")
	return
}

func genFilename() string {
	t := time.Now()
	return fmt.Sprintf("%s.zip", t.Format("20060102T150405.000"))
}

func localConnIP(conn net.Conn) net.IP {
	ipstr, _ := lldars.ParseIpPort(conn.LocalAddr().String())
	return net.ParseIP(ipstr).To4()
}

func Error(_err error) {
	if _err != nil {
		log.Panic(_err)
	}
}
