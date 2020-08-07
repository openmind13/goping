package ping

import (
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

const defaultReadDeadline = 10 * time.Second

// Pinger struct
type Pinger struct {
	localAddr    net.Addr
	targetAddr   net.Addr
	readDeadLine time.Duration
	responseTime time.Duration
	message      icmp.Message
	binMessage   []byte
}

// NewPinger - create new pinger
func NewPinger() Pinger {
	localAddr, err := getLocalAddr()
	if err != nil {
		panic(err)
	}

	message := icmp.Message{
		Type: ipv4.ICMPTypeEcho,
		Code: 0,
		Body: &icmp.Echo{
			ID:   os.Getpid() & 0xffff,
			Seq:  1,
			Data: []byte("DEFAULT-MESSAGE"),
		},
	}
	binMessage, err := message.Marshal(nil)
	if err != nil {
		panic(err)
	}

	return Pinger{
		localAddr:    localAddr,
		targetAddr:   nil,
		readDeadLine: defaultReadDeadline,
		responseTime: 0,
		message:      message,
		binMessage:   binMessage,
	}
}

// Ping - send icmp echo requests to target
func (pinger *Pinger) Ping(targetAddr string) {
	//pinger.targetAddr = ipAddr.String()

	strings.TrimRight(pinger.localAddr.String(), "/20")
	fmt.Println(pinger.localAddr)
	conn, err := icmp.ListenPacket("ip4:icmp", pinger.localAddr.String())
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	start := time.Now()

	// dstAddr, err := net.ResolveIPAddr("ip4", pinger.localAddr.String())
	// if err != nil {
	// 	panic(err)
	// }

	n, err := conn.WriteTo(pinger.binMessage, pinger.targetAddr)
	if err != nil {
		panic(err)
	}

	replyBuffer := make([]byte, 1500)
	err = conn.SetDeadline(time.Now().Add(pinger.readDeadLine))
	if err != nil {
		panic(err)
	}

	n, peer, err := conn.ReadFrom(replyBuffer)
	if err != nil {
		panic(err)
	}

	pinger.responseTime = time.Since(start)

	replyMessage, err := icmp.ParseMessage(1, replyBuffer[:n])
	if err != nil {
		panic(err)
	}

	switch replyMessage.Type {
	case ipv4.ICMPTypeEchoReply:
		fmt.Printf("echo reply from %v\n", peer)
		fmt.Println(replyMessage)
	default:
		fmt.Println("have not a reply message")
		fmt.Println(replyMessage)
	}
}

// SetReadDeadline - set read deadline
func (pinger *Pinger) SetReadDeadline(duration time.Duration) {
	pinger.readDeadLine = duration
}
