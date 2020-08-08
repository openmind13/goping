package main

import (
	"fmt"
	"os"
	"os/signal"

	"github.com/openmind13/goping/ping"
)

const (
	localAddr = "172.23.177.246"

	targetIP       = "192.168.0.10"
	targetIPGoogle = "google.com"
)

func main() {
	fmt.Printf("======\n")

	pinger, err := ping.NewPinger(targetIPGoogle)
	if err != nil {
		panic(err)
	}

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)

	go catchSignal(ch)

	err = pinger.Ping()
	if err != nil {
		panic(err)
	}

}

func catchSignal(ch chan os.Signal) {
	switch <-ch {
	case os.Interrupt:
		fmt.Printf("\nprint statistics\n")
		os.Exit(0)
	default:

	}
}
