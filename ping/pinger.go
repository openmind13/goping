package ping

import (
	"fmt"
	"net"
	"os"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

const (
	defaultTimeout = 10 * time.Second

	protocolICMP     = 1
	protocolIPv6ICMP = 58

	defaultLocalAddr = "172.23.188.146"

	targetIP  = "192.168.0.10"
	targetIP2 = "google.com"
)

var (
	protocolIPv4 = map[string]string{"ip": "ip4:icmp", "udp": "udp4"}
	protocolIPv6 = map[string]string{"ip": "ip6:ipv6-icmp", "udp": "udp6"}
)

// Pinger struct
type Pinger struct {
	timeout        time.Duration
	interval       time.Duration
	localAddr      string
	targetAddr     string
	rawTargetAddr  string
	transportLayer string
	targetIPAddr   net.IPAddr
	count          int
}

// NewPinger - create new pinger
func NewPinger(targetAddr string) (*Pinger, error) {
	ipAddr, err := net.ResolveIPAddr("ip", targetAddr)
	if err != nil {
		return nil, err
	}

	pinger := Pinger{
		timeout:       defaultTimeout,
		interval:      1 * time.Second,
		localAddr:     defaultLocalAddr,
		targetAddr:    ipAddr.String(),
		rawTargetAddr: targetAddr,
		targetIPAddr:  *ipAddr,
		count:         -1,
	}

	return &pinger, nil
}

// SetCount - set count of sending packets
func (pinger *Pinger) SetCount(count int) {
	pinger.count = count
}

// SetTimeout - set timeout for ping
func (pinger *Pinger) SetTimeout(duration time.Duration) {
	pinger.timeout = duration
}

// Ping - start ping
func (pinger *Pinger) Ping() error {

	body := &icmp.Echo{
		ID:   os.Getpid() & 0xffff,
		Seq:  1,
		Data: []byte("DEFAULT-MESSAGE"),
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

	conn, err := icmp.ListenPacket("ip4:icmp", pinger.localAddr)
	if err != nil {
		fmt.Println("error in icmp.ListenPacket()")
		return err
	}
	err = conn.SetDeadline(time.Now().Add(pinger.timeout))
	if err != nil {
		return err
	}

	fmt.Printf("PING (%v)\n", pinger.rawTargetAddr)
	i := pinger.count
	for i < 0 {

		start := time.Now()
		_, err := conn.WriteTo(binaryMessage, &pinger.targetIPAddr)
		if err != nil {
			return err
		}

		//fmt.Printf("send %d bytes to (%s)\n", byteCount, pinger.targetAddr)

		recvByteCount, peer, err := conn.ReadFrom(replyBuffer)
		if err != nil {
			return err
		}

		responseTime := time.Since(start)
		replyMsg, err := icmp.ParseMessage(protocolICMP, replyBuffer[:recvByteCount])
		if err != nil {
			return err
		}

		switch replyMsg.Type {
		case ipv4.ICMPTypeEchoReply:
			fmt.Printf("%d bytes from (%s) time=%v\n", recvByteCount, peer, responseTime)
		default:
			fmt.Println("have not a reply message")
			fmt.Println(replyMsg)
		}

		time.Sleep(pinger.interval)
	}

	return nil
}
