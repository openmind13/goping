package ping

import (
	"errors"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strings"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

const (
	defaultLocalAddr = "172.23.187.32"
)

// Protocol Constants
const (
	protocolICMP     = 1
	protocolIPv6ICMP = 58

	ip4icmp  = "ip4:icmp"
	ip6icmp  = "ip6:ipv6-icmp"
	udp4icmp = "upd4"
	udp6icmp = "udp6"
)

// Errors
var (
	errLocalAddrNotFound = errors.New("Local addr not found")
)

// Pinger struct
type Pinger struct {
	interval time.Duration
	pingTime time.Duration

	localAddr      string
	targetAddr     string
	rawTargetAddr  string
	transportLayer string
	targetIPAddr   net.IPAddr
	count          int

	sendedPackets   int
	recievedPackets int

	channel chan os.Signal
}

// NewPinger - create new pinger
func NewPinger(targetAddr string) (*Pinger, error) {
	ipAddr, err := net.ResolveIPAddr("ip", targetAddr)
	if err != nil {
		return nil, err
	}

	pinger := Pinger{
		interval:      1000 * time.Millisecond,
		localAddr:     defaultLocalAddr,
		targetAddr:    ipAddr.String(),
		rawTargetAddr: targetAddr,
		targetIPAddr:  *ipAddr,
		count:         -1,
	}

	// err = pinger.getLocalAddr()
	// if err != nil {
	// 	return nil, err
	// }

	pinger.setInformChan()

	return &pinger, nil
}

func (pinger *Pinger) setInformChan() {
	pinger.channel = make(chan os.Signal, 1)
	signal.Notify(pinger.channel, os.Interrupt)
}

func (pinger *Pinger) catchSignal() {
	switch <-pinger.channel {
	case os.Interrupt:
		fmt.Printf("\n--- %s ping statistics ---\n", pinger.rawTargetAddr)
		fmt.Printf("%d packets transmitted, %d received\n",
			pinger.sendedPackets,
			pinger.recievedPackets)

		pinger.sendedPackets = 0
		pinger.recievedPackets = 0

		os.Exit(0)
	}
}

func (pinger *Pinger) getLocalAddr() error {
	netInterfaces, err := net.Interfaces()
	if err != nil {
		return err
	}
	for _, i := range netInterfaces {
		if strings.Contains(i.Flags.String(), "up") &&
			strings.Contains(i.Flags.String(), "broadcast") &&
			strings.Contains(i.Flags.String(), "multicast") {

			ip, err := i.Addrs()
			if err != nil {
				return err
			}

			ipAddr, err := convertToIPAddr(ip[0])
			if err != nil {
				return err
			}
			pinger.localAddr = ipAddr.String()
			return nil
		}
	}

	return errLocalAddrNotFound
}

// Ping - start ping
func (pinger *Pinger) Ping() error {
	pinger.recievedPackets = 0
	pinger.sendedPackets = 0

	go pinger.catchSignal()

	body := &icmp.Echo{
		ID:   os.Getpid() & 0xffff,
		Seq:  1,
		Data: []byte("DEFAULT-MESSAGE-HELLO-WORLD-FROM-GOLANG-PING"),
	}

	message := icmp.Message{
		Type: ipv4.ICMPTypeEcho,
		Code: 0,
		Body: body,
	}

	binaryMessage, err := message.Marshal(nil)
	if err != nil {
		return nil
	}

	replyBuffer := make([]byte, 512)

	conn, err := icmp.ListenPacket(ip4icmp, pinger.localAddr)
	if err != nil {
		fmt.Println("error in icmp.ListenPacket()")
		return err
	}

	// set deadline for max duration for ping
	// err = conn.SetDeadline(time.Now().Add(pinger.timeout))
	// if err != nil {
	// 	return err
	// }

	fmt.Printf("PING (%v) %d bytes of data\n", pinger.rawTargetAddr, len(binaryMessage))
	i := pinger.count

	for i < 0 {
		start := time.Now()
		_, err := conn.WriteTo(binaryMessage, &pinger.targetIPAddr)
		if err != nil {
			fmt.Printf("Error in WriteTo()\n")
			return err
		}
		pinger.sendedPackets++

		//fmt.Printf("send %d bytes to (%s)\n", byteCount, pinger.targetAddr)

		recvByteCount, peer, err := conn.ReadFrom(replyBuffer)
		if err != nil {
			return err
		}
		pinger.recievedPackets++

		responseTime := time.Since(start)
		replyMsg, err := icmp.ParseMessage(protocolICMP, replyBuffer[:recvByteCount])
		if err != nil {
			return err
		}

		switch replyMsg.Type {
		case ipv4.ICMPTypeEchoReply:
			fmt.Printf("%d bytes from (%s) time=%v\n",
				recvByteCount,
				peer,
				responseTime)
		default:
			fmt.Println("have not a reply message")
			fmt.Println(replyMsg)
		}

		time.Sleep(pinger.interval)
	}

	return nil
}

// SetCount - set count of sending packets
func (pinger *Pinger) SetCount(count int) {
	pinger.count = count
}

// SetInterval - set timeout for ping
func (pinger *Pinger) SetInterval(duration time.Duration) {
	pinger.interval = duration
}
