package main

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	"net"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/silmin/lldars/pkg/lldars"
)

const (
	IntervalSeconds        = 1
	TimeoutSeconds         = 10
	ExpirationSecondsOfAck = 30
	LLDARSObjectPath       = "./send_data/"
	BackupObjectsPath      = "./backups/"
)

type BcType uint8

const (
	Limited BcType = iota
	Directed
)

type Server struct {
	Id           uint32
	Mode         lldars.LLDARSServeMode
	DstNetwork   net.IPNet
	BcType       BcType
	ClientBCAddr string
	ServerBCAddr string
}

func NewServer(id uint32, mode lldars.LLDARSServeMode, target string) (*Server, error) {
	s := &Server{
		Id:   id,
		Mode: mode,
	}

	if s.Id == 0 {
		s.Id = uuid.New().ID()
	}

	if target != "255.255.255.255" {
		_, ipnet, err := net.ParseCIDR(target)
		if err != nil {
			return &Server{}, errors.New("non-valid dst IP")
		}
		bc, err := bcAddr(ipnet)
		if err != nil {
			return &Server{}, errors.New("non-valid dst IP")
		}
		s.BcType = Directed
		s.ClientBCAddr = joinIpPort(bc.String(), strconv.Itoa(lldars.ClientBCPort))
		s.ServerBCAddr = joinIpPort(bc.String(), strconv.Itoa(lldars.ServerBCPort))
	} else {
		s.BcType = Limited
		s.ClientBCAddr = joinIpPort(net.IPv4bcast.String(), strconv.Itoa(lldars.ClientBCPort))
		s.ServerBCAddr = joinIpPort(net.IPv4bcast.String(), strconv.Itoa(lldars.ServerBCPort))
	}

	return s, nil
}

func (s Server) Serve(ctx context.Context) {
	log.Printf("Server ID: %v\n", s.Id)

	if s.Mode == lldars.RevivalMode {
		syCtx, syClose := context.WithCancel(ctx)
		defer syClose()
		s.syncObjects(syCtx)
	}
	bcCCtx, bcCClose := context.WithCancel(ctx)
	bcSCtx, bcSClose := context.WithCancel(ctx)
	brCtx, brClose := context.WithCancel(ctx)
	defer func() {
		brClose()
		bcCClose()
		bcSClose()
	}()

	ackClientCache := NewIdCache()
	ackServerCache := NewIdCache()
	backupCache := NewIdCache()

	go s.listenDiscoverBroadcast(bcCCtx, joinIpPort(net.IPv4zero.String(), strconv.Itoa(lldars.ClientBCPort)), ackClientCache)
	go s.listenDiscoverBroadcast(bcSCtx, joinIpPort(net.IPv4zero.String(), strconv.Itoa(lldars.ServerBCPort)), ackServerCache)
	go s.backupRegularly(brCtx)
	s.listenService(backupCache)

	return
}

func (s Server) listenService(cache *IdCache) {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", lldars.ServicePort))
	Error(err)
	for {
		conn, err := ln.Accept()
		Error(err)
		go handleService(conn, s.Id, cache)
	}
}

func handleService(conn net.Conn, serverId uint32, cache *IdCache) {
	defer conn.Close()
	rl := ReadLLDARSHeader(conn)
	log.Printf("Receive from: %v\n", rl.Origin)

	if rl.Type == lldars.GetObjectRequest {
		_ = ReadLLDARSPayload(conn, rl.Length)

		sendObjects(conn, serverId, LLDARSObjectPath)
	} else if rl.Type == lldars.SyncObjectRequest {
		_ = ReadLLDARSPayload(conn, rl.Length)

		sendSyncObjects(conn, rl, serverId)
	} else if rl.Type == lldars.BackupObjectRequest {
		_ = ReadLLDARSPayload(conn, rl.Length)

		receiveBackupObjects(conn, rl, serverId, cache)
	}

	return
}

func (s Server) listenDiscoverBroadcast(ctx context.Context, listenAddr string, cache *IdCache) {
	udpAddr, err := net.ResolveUDPAddr("udp", listenAddr)
	Error(err)
	udpLn, err := net.ListenUDP("udp", udpAddr)
	Error(err)
	defer udpLn.Close()

	origin := localConnIP(udpLn).String()

	log.Printf("Listened Delivering Requests *:* udp > %s\n", listenAddr)
	for {
		buf := make([]byte, lldars.LLDARSLayerSize+len(lldars.DiscoverBroadcastPayload))
		l, err := udpLn.Read(buf)
		Error(err)
		msg := buf[:l]
		rl := lldars.Unmarshal(msg)
		log.Printf("Receive BC from: %v\n", rl.Origin)

		if rl.Type == lldars.DiscoverBroadcast && rl.Origin.String() != origin {
			if rl.ServerId != 0 {
				if !cache.Exists(cacheAckKey(s.Id)) {
					cache.Push(cacheAckKey(s.Id), rl.ServerId, time.Now().Add(ExpirationSecondsOfAck*time.Second).UnixNano())
					ackBroadcast(s.Id, rl, origin)
				}
			} else {
				ackBroadcast(s.Id, rl, origin)
			}
		}
	}
}

func ackBroadcast(serverId uint32, rl lldars.LLDARSLayer, origin string) {
	sl := lldars.NewServerPortNotify(serverId, net.ParseIP(origin), lldars.ServicePort)
	ipp := fmt.Sprintf("%s:%d", rl.Origin.String(), rl.ServicePort)
	conn, err := net.Dial("tcp", ipp)
	Error(err)
	defer conn.Close()
	conn.Write(sl.Marshal())
	log.Printf("Ack to: %v\tmsg: %s\n", conn.RemoteAddr().String(), sl.Payload)
}

func localConnIP(conn net.Conn) net.IP {
	ipstr, _ := lldars.ParseIpPort(conn.LocalAddr().String())
	return net.ParseIP(ipstr).To4()
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

func cacheAckKey(serverId uint32) string {
	return fmt.Sprintf("%v-ack", serverId)
}

func Error(_err error) {
	if _err != nil {
		log.Panic(_err)
	}
}
