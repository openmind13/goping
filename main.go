package main

import (
	"fmt"
	"os"

	"github.com/openmind13/goping/ping"
)

const (
	localAddr = "172.23.177.246"

	gatewayIP = "192.168.0.10"
)

func main() {
	// fmt.Printf("======\n")

	if len(os.Args) < 2 {
		fmt.Printf("enter IP or domain name\n")
		os.Exit(0)
	} else {
		fmt.Printf("%s\n", os.Args[1])
	}

	pinger, err := ping.NewPinger(os.Args[1])
	if err != nil {
		panic(err)
	}

	err = pinger.Ping()
	if err != nil {
		panic(err)
	}
}
