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
	deadline time.Duration

	startPingTime         time.Time
	lastSendingPacketTime time.Time

	stringLocalAddr string
	localAddr       *net.IPAddr

	stringTargetAddr    string
	rawStringTargetAddr string
	targetAddr          *net.IPAddr

	count int

	message       icmp.Message
	binaryMessage []byte

	sendedPackets   int
	recievedPackets int

	replyBuffer [512]byte

	conn *icmp.PacketConn

	signalChannel chan os.Signal
}

// NewPinger - create new pinger
func NewPinger(targetAddr string) (*Pinger, error) {
	ipAddr, err := net.ResolveIPAddr("ip", targetAddr)
	if err != nil {
		return nil, err
	}
	pinger := Pinger{
		interval:            1000 * time.Millisecond,
		stringTargetAddr:    ipAddr.String(),
		rawStringTargetAddr: targetAddr,
		targetAddr:          ipAddr,
		count:               -1, // -1 for ping in infinity loop
		replyBuffer:         [512]byte{},
	}
	if err = pinger.setLocalAddr(); err != nil {
		return nil, err
	}
	pinger.setSignalChan()
	if err = pinger.setMessage(); err != nil {
		return nil, err
	}
	return &pinger, nil
}

// setInformChannel ...
func (pinger *Pinger) setSignalChan() {
	pinger.signalChannel = make(chan os.Signal, 1)
	signal.Notify(pinger.signalChannel, os.Interrupt)
}

// catchExitSignal listen pinger.signalChannel
func (pinger *Pinger) catchExitSignal() {
	switch <-pinger.signalChannel {
	case os.Interrupt:
		loss := (1 - (pinger.recievedPackets / pinger.sendedPackets)) * 100
		fmt.Printf("\n--- %s ping statistics ---\n", pinger.stringTargetAddr)
		fmt.Printf("%d packets transmitted, %d received, %v%% packet loss, time %v\n",
			pinger.sendedPackets,
			pinger.recievedPackets,
			loss,
			time.Since(pinger.startPingTime),
		)
		fmt.Printf("=========================================\n")
		pinger.sendedPackets = 0
		pinger.recievedPackets = 0
		os.Exit(0)
	}
}

// Ping - start ping
func (pinger *Pinger) Ping() error {
	pinger.recievedPackets = 0
	pinger.sendedPackets = 0
	conn, err := icmp.ListenPacket(ip4icmp, pinger.localAddr.String())
	if err != nil {
		fmt.Println("error in icmp.ListenPacket()")
		return err
	}
	pinger.conn = conn
	defer pinger.conn.Close()
	go pinger.catchExitSignal()
	pinger.startPingTime = time.Now()
	go pinger.recvMessages()
	if err = pinger.sendMessages(); err != nil {
		return err
	}
	return nil
}

// recvMessages recv messages until <-pinger.signalChannel
func (pinger *Pinger) recvMessages() error {
	for {
		select {
		case <-pinger.signalChannel:
			os.Exit(0)
		default:
			recvByteCount, peer, err := pinger.conn.ReadFrom(pinger.replyBuffer[:])
			if err != nil {
				return err
			}
			pinger.recievedPackets++
			replyMsg, err := icmp.ParseMessage(protocolICMP, pinger.replyBuffer[:recvByteCount])
			if err != nil {
				return err
			}
			switch replyMsg.Type {
			case ipv4.ICMPTypeEchoReply:
				fmt.Printf("%d bytes from (%s) icmp_seq=%d time=%v\n",
					recvByteCount,
					peer,
					pinger.sendedPackets,
					time.Since(pinger.lastSendingPacketTime),
				)
			default:
				fmt.Println("have not a reply message")
				fmt.Println(replyMsg)
			}
		}
	}
}

// sendMessages start send messages until not <-pinger.signalChannel
func (pinger *Pinger) sendMessages() error {
	fmt.Printf("=========================================\n")
	fmt.Printf("PING (%s) send %d bytes of data\n",
		pinger.rawStringTargetAddr,
		len(pinger.binaryMessage))
	for {
		select {
		case <-pinger.signalChannel:
			return nil
		default:
			pinger.lastSendingPacketTime = time.Now()
			if _, err := pinger.conn.WriteTo(pinger.binaryMessage, pinger.targetAddr); err != nil {
				fmt.Printf("Error in WriteTo()\n")
				return err
			}
			// _, err := conn.WriteTo(pinger.binaryMessage, &pinger.targetIPAddr)
			pinger.sendedPackets++
			time.Sleep(pinger.interval)
		}
	}
}

// SetCount - set count of sending packets
func (pinger *Pinger) SetCount(count int) {
	pinger.count = count
}

// SetInterval - set timeout for ping
func (pinger *Pinger) SetInterval(duration time.Duration) {
	pinger.interval = duration
}

// SetDeadline ...
func (pinger *Pinger) SetDeadline(duration time.Duration) error {
	if err := pinger.conn.SetDeadline(time.Now().Add(pinger.deadline)); err != nil {
		return err
	}
	return nil
}

// setMessage ...
func (pinger *Pinger) setMessage() error {
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
		return err
	}
	pinger.message = message
	pinger.binaryMessage = binaryMessage
	return nil
}

// setLocalAddr for finding local add on interface
func (pinger *Pinger) setLocalAddr() error {
	netInterfaces, err := net.Interfaces()
	if err != nil {
		return err
	}
	for _, i := range netInterfaces {
		if strings.Contains(i.Flags.String(), "up") &&
			strings.Contains(i.Flags.String(), "broadcast") &&
			strings.Contains(i.Flags.String(), "multicast") {
			ipaddrs, err := i.Addrs()
			if err != nil {
				return err
			}
			for _, addr := range ipaddrs {
				switch v := addr.(type) {
				case *net.IPAddr:
					// fmt.Println(v.IP)
				case *net.IPNet:
					ip := v.IP
					pinger.localAddr = &net.IPAddr{
						IP:   ip,
						Zone: "",
					}
					return nil
				}
			}
			return nil
		}
	}
	return errLocalAddrNotFound
}
