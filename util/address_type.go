package util

import "net"

// AddressType denotes the type of IP address.
type AddressType int

const (
	// IP4 represents IPv4 address.
	IP4 AddressType = iota
	// IP6 represents IPv6 address.
	IP6 AddressType = iota
)

// GetAddressType returns the AddressType of the given IP.
func GetAddressType(ipAddr net.IP) AddressType {
	if ipAddr.To4() != nil {
		return IP4
	}

	return IP6
}
