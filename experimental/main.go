package main

import (
	"context"

	"github.com/silmin/lldars/experimental/client"
	"github.com/silmin/lldars/experimental/server"
)

func main() {
	// Client
	c := client.NewClient(1000)
	ctx := context.Background()
	c.Broadcast(ctx, "192.168.100.255:60000", ":60000", "192.168.100.1:60000")

	// Server
	server.Server("192.168.100.1:60000")
}
