package ping

import (
	"errors"
	"net"
	"strings"
)

func getLocalAddr() (*net.IPAddr, error) {
	netInterfaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	for _, i := range netInterfaces {
		if strings.Contains(i.Flags.String(), "up") &&
			strings.Contains(i.Flags.String(), "broadcast") &&
			strings.Contains(i.Flags.String(), "multicast") {

			ip, err := i.Addrs()
			if err != nil {
				return nil, err
			}

			ipAddr, err := convertToIPAddr(ip[0])
			return ipAddr, nil
		}
	}

	return nil, errors.New("Local addr not found")
}

func convertToIPAddr(addr net.Addr) (*net.IPAddr, error) {
	stringAddr := addr.String()
	ipAddr, err := net.ResolveIPAddr("ip", stringAddr)
	if err != nil {
		return nil, err
	}
	return ipAddr, nil
}

func countMessageSize(msg []byte) int {
	return len(msg)
}
