package main

import (
	"context"
	"encoding/binary"
	"errors"
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
	ReceiveObjectPath = "./receive_data/"
)

type Client struct {
	BcDst net.IP
}

func NewClient(bcd string) (*Client, error) {
	if bcd != "255.255.255.255" {
		_, ipnet, err := net.ParseCIDR(bcd)
		if err != nil {
			return &Client{}, errors.New("non-valid dst IP")
		}
		ip, err := bcAddr(ipnet)
		if err != nil {
			return &Client{}, errors.New("non-valid dst IP")
		}

		return &Client{
			BcDst: ip,
		}, nil
	} else {
		return &Client{
			BcDst: net.ParseIP(bcd),
		}, nil
	}
}

func (c *Client) DoAct() {
	ctx, close := context.WithTimeout(context.Background(), time.Duration(TimeoutSeconds)*time.Second)
	defer close()

	serviceAddr := make(chan string)
	go DiscoverBroadcast(ctx, 0, serviceAddr, joinIpPort(c.BcDst.String(), strconv.Itoa(lldars.ClientBCPort)))

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
	ip := net.ParseIP(ipstr).To4()
	if ip == nil {
		return net.ParseIP("0.0.0.0").To4()
	} else {
		return ip
	}
}

func remoteConnIP(conn net.Conn) net.IP {
	ipstr, _ := lldars.ParseIpPort(conn.RemoteAddr().String())
	ip := net.ParseIP(ipstr).To4()
	if ip == nil {
		return net.ParseIP("0.0.0.0").To4()
	} else {
		return ip
	}
}

func bcAddr(n *net.IPNet) (net.IP, error) {
	if n.IP.To4() == nil {
		return net.IP{}, errors.New("does not support IPv6 addresses.")
	}
	ip := make(net.IP, len(n.IP.To4()))
	binary.BigEndian.PutUint32(ip, binary.BigEndian.Uint32(n.IP.To4())|^binary.BigEndian.Uint32(net.IP(n.Mask).To4()))
	return ip, nil
}

func joinIpPort(ip string, port string) string {
	if port[0] == ':' {
		port = port[1:]
	}
	return ip + ":" + port
}

func Error(_err error) {
	if _err != nil {
		log.Panic(_err)
	}
}
