package main

import (
	"context"

	"github.com/silmin/lldars/pkg/lldars"
)

func main() {
	ctx, close := context.WithCancel(context.Background())
	defer close()
	Server(ctx, "192.168.100.255:60000", "192.168.100.2", 0, lldars.NormalMode)
}
