package fluid

import "net"

// ExternalIP looks up an the first available external IP address. It is used
// by the LocalReplica function to automatically instantiate a local node.
func ExternalIP() (string, error) {

	// Get addresses for the interface
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", NetworkError("could not get interface addresses", err)
	}

	// Go through each address to find a an IPv4
	for _, addr := range addrs {

		var ip net.IP

		switch val := addr.(type) {
		case *net.IPNet:
			ip = val.IP
		case *net.IPAddr:
			ip = val.IP
		}

		if ip == nil || ip.IsLoopback() {
			continue // ignore loopback and nil addresses
		}

		ip = ip.To4()
		if ip == nil {
			continue // not an ipv4 address
		}

		return ip.String(), nil
	}

	return "", NetworkError("could not get external addr, requires network connection", nil)
}

// ResolveAddr accepts an address as a string and if the IP address is missing
// it replaces it with the result from ExternalIP then returns the addr
// string. Likewise if the Port is missing, it returns an address with the
// DefaultPort appended to the address string.
func ResolveAddr(addr string) (string, error) {

	tcpAddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return "", NetworkError("could not resolve address", err)
	}

	if tcpAddr.IP == nil {
		ipstr, err := ExternalIP()
		if err != nil {
			return "", err
		}

		tcpAddr.IP = net.ParseIP(ipstr)
	}

	if tcpAddr.Port == 0 {
		tcpAddr.Port = DefaultPort
	}

	return tcpAddr.String(), nil
}
