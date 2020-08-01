package main

import (
	"fmt"
	"log"
	"net"
	"os"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

const targetIP = "localhost"

func main() {
	// byteBuffer := make([]byte, 10)

	// byteBuffer[7] = 0x24
	// byteBuffer[2] = 0x15
	// byteBuffer[0] = 0x1
	// byteBuffer[5] = 0x56

	// for i := range byteBuffer {
	// 	fmt.Print(i, " >\t")
	// 	fmt.Println(byteBuffer[i], "\t>\t", string(byteBuffer[i]))
	// 	fmt.Println("-----------------------------------------")
	// }
	// fmt.Println(byteBuffer)
	// fmt.Println(string(byteBuffer))

	// first argument network (string) - udp4 - operation not permitted by OS (Linux WSL)
	conn, err := icmp.ListenPacket("ip4:icmp", "localhost")
	if err != nil {
		log.Fatalf("listen err, %s", err)
	}
	defer conn.Close()

	fmt.Println("successfully open socket")

	// fmt.Printf("\nConnection:\n%v\n", conn)

	//fmt.Println("debug message")

	wMessage := icmp.Message{
		Type: ipv4.ICMPTypeEcho,
		Code: 8, // Echo type
		Body: &icmp.Echo{
			ID:   os.Getpid() & 0xffff,
			Seq:  1,
			Data: []byte("HELLO-FROM-PIROGOV-ALEXANDER"),
		},
	}
	// fmt.Printf("wEchoMessage\n%v\n%v\n", wEchoMessage, wEchoMessage.Body)

	wBytesMessage, err := wMessage.Marshal(nil)
	if err != nil {
		panic(err)
	}

	nbytes, err := conn.WriteTo(wBytesMessage, &net.IPAddr{IP: net.ParseIP(targetIP)})
	if err != nil {
		panic(err)
	}
	fmt.Println("\nwrite to net.Conn icmp byte packet")
	fmt.Printf("writing bytes: %v\n", nbytes)

	// make byte buffer to store data from connection
	byteBuffer := make([]byte, 1500)
	_, peer, err := conn.ReadFrom(byteBuffer)
	if err != nil {
		panic(err)
	}
	fmt.Printf("\npeer: %v\n\n", peer)

	rMessage, err := icmp.ParseMessage(ipv4.ICMPTypeEchoReply.Protocol(), byteBuffer)
	if err != nil {
		panic(err)
	}

	switch rMessage.Type {
	case ipv4.ICMPTypeEchoReply:
		fmt.Printf("got reflection from %v", peer)

	default:
		fmt.Printf("got %+v; want echo reply", rMessage)
	}

	fmt.Println()
}
