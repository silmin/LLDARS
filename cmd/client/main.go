package main

import "flag"

func main() {
	f := flag.String("dst", "255.255.255.255", "destination network address")
	flag.Parse()
	c, err := NewClient(*f)
	if err != nil {
		Error(err)
	}
	c.DoAct()
}
