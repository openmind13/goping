package main

import "github.com/openmind13/goping/ping"

const targetIP = "192.168.0.10"
const targetIP2 = "google.com"

func main() {
	p := ping.NewPinger()
	p.Ping(targetIP)

	// addr, time, err := ping.Ping(targetIP2)
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println(addr)
	// fmt.Println(time)

	// addr, err := ping.GetLocalAddr()
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println(addr)
	// fmt.Println(addr.Network())
	// address := addr.String()
	// fmt.Println(address)
}
