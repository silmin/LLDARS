package main

import (
	"github.com/silmin/lldars/client"
)

func main() {
	// Client
	c := client.NewClient()
	c.DoAct()

	// Server
	// ctx, close := context.WithTimeout(context.Background(), time.Duration(TimeoutSeconds)*time.Second)
	// server.Server(ctx, "192.168.100.255:60000", "192.168.100.2", lldars.NormalMode)
}
