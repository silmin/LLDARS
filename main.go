package main

import (
	"github.com/silmin/lldars/client"
)

func main() {
	// Client
	c := client.NewClient(1000)
	c.DoAct()

	// Server
	//server.Server("192.168.100.255:60000", "192.168.100.2")
}
