package main

import (
	"context"
	"time"

	"github.com/silmin/lldars/experimental/client"
)

func main() {
	// Client
	c := client.NewClient(1000)
	ctx, close := context.WithTimeout(context.Background(), time.Duration(client.TimeoutSeconds)*time.Second)
	c.Broadcast(ctx, "192.168.100.255:60000")
	close()

	// Server
	//server.Server("192.168.100.255:60000")
}
