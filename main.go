package main

import (
	"github.com/silmin/lldars/client"
)

func main() {
	// Client
	c := client.NewClient()
	c.DoAct()

	// Server
	// ctx, close := context.WithCancel(context.Background())
	// defer close()
	// server.Server(ctx, "192.168.100.255:60000", "192.168.100.2", 0, lldars.NormalMode)
}
