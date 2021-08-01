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
	go c.discoverBroadcast(ctx, serviceAddr)

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
		buf := make([]byte, lldars.LLDARSLayerSize)
		l, err := conn.Read(buf)
		Error(err)
		rl := lldars.Unmarshal(buf[:l])
		log.Printf("Recieve from: %v\tpayload-len: %d\n", rl.Origin, rl.Length)
		log.Printf("serverId: %d\n", rl.ServerId)
		if rl.Type == lldars.EndOfDelivery {
			sl := lldars.NewReceivedObjects(0, localIP(conn), 0)
			conn.Write(sl.Marshal())
			log.Println("--End receiving objects--")
			break
		} else if rl.Type != lldars.DeliveryObject {
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
			log.Printf("Read Parts %d (%d/%d)\n", l, len(obj), rl.Length)
		}

		if len(obj) != 0 {
			err = ioutil.WriteFile(filename, obj, 0644)
			Error(err)
			log.Printf("Read Object > %s, len: %d\n", filename, rl.Length)
		}
	}

	log.Println("End getObjects()")
	return
}

func (c *Client) discoverBroadcast(ctx context.Context, serviceAddr chan<- string) {
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
	go c.listenAck(listenCtx, udpLn, serviceAddr)

	addr, port := lldars.ParseIpPort(udpLn.LocalAddr().String())
	Error(err)
	p, _ := strconv.Atoi(port)
	l := lldars.NewDiscoverBroadcast(0, net.ParseIP(addr).To4(), uint16(p))
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

func (c *Client) listenAck(ctx context.Context, udpLn *net.UDPConn, serviceAddr chan<- string) {
	log.Println("Start listenAck()")

	buf := make([]byte, lldars.LLDARSLayerSize+len(lldars.ServicePortNotifyPayload))
	l, err := udpLn.Read(buf)
	Error(err)
	msg := string(buf[:l])
	rl := lldars.Unmarshal([]byte(msg))
	log.Println(rl)

	log.Printf("Recieve from: %v\tmsg: %s\n", rl.Origin, rl.Payload)

	serviceAddr <- fmt.Sprintf("%s:%d", rl.Origin.String(), rl.ServicePort)
}

func genFilename() string {
	t := time.Now()
	return fmt.Sprintf("%s.zip", t.Format("20060102T150405.000"))
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
