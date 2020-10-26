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
	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {
	if len(os.Args) < 2 {
		fmt.Printf("enter IP or domain name\n")
		os.Exit(0)
	}
	pinger, err := ping.NewPinger(os.Args[1])
	if err != nil {
		return err
	}
	if err = pinger.Ping(); err != nil {
		return err
	}
	return nil
}
