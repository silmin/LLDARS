package main

import (
	"context"
	"flag"

	"github.com/silmin/lldars/pkg/lldars"
)

func main() {
	f := flag.String("dst", "255.255.255.255", "destination network address")
	flag.Parse()

	s, err := NewServer(0, lldars.NormalMode, *f)
	if err != nil {
		Error(err)
	}

	ctx, close := context.WithCancel(context.Background())
	defer close()
	s.Serve(ctx)
}
