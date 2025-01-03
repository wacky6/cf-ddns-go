package util

import (
	"fmt"
	"net"
	"sort"
)

var privateCidrs = []*net.IPNet{
	cidr("10.0.0.0/8"),     // RFC 1918 IPv4 private network address
	cidr("100.64.0.0/10"),  // RFC 6598 IPv4 carrier NAT address
	cidr("127.0.0.0/8"),    // RFC 1122 IPv4 loopback address
	cidr("169.254.0.0/16"), // RFC 3927 IPv4 link local address
	cidr("172.16.0.0/12"),  // RFC 1918 IPv4 private network address
	cidr("192.0.0.0/24"),   // RFC 6890 IPv4 IANA address
	cidr("192.0.2.0/24"),   // RFC 5737 IPv4 documentation address
	cidr("192.168.0.0/16"), // RFC 1918 IPv4 private network address
	cidr("::1/128"),        // RFC 1884 IPv6 loopback address
	cidr("fe80::/10"),      // RFC 4291 IPv6 link local addresses
	cidr("fc00::/7"),       // RFC 4193 IPv6 unique local addresses
	cidr("fec0::/10"),      // RFC 1884 IPv6 site-local addresses
	cidr("2001:db8::/32"),  // RFC 3849 IPv6 documentation address
}

func cidr(s string) *net.IPNet {
	_, block, err := net.ParseCIDR(s)
	if err != nil {
		panic(fmt.Sprintf("Bad CIDR %s: %s", s, err))
	}
	return block
}

func contains(arr []string, target string) bool {
	for _, item := range arr {
		if item == target {
			return true
		}
	}

	return false
}

func isPrivate(ip net.IP) bool {
	for _, privateCidr := range privateCidrs {
		if privateCidr.Contains(ip) {
			return true
		}
	}

	return false
}

// DetectAddress picks the appropriate IP detection method based on `detector`.
func DetectAddress(detector string, addrType AddressType, ifaceNames []string) (string, error) {
	switch detector {
	case "iface":
		return DetectByInterface(addrType, ifaceNames)
	default:
		return "", fmt.Errorf("detector %q is not supported", detector)
	}
}

// DetectByInterface finds the most-likely public IP address as a string, by
// inspecting addresses assigned to OS network interfaces. If ifaceNames are
// not empty, it will filter ip addresses based on the associated interface
// name.
func DetectByInterface(addrType AddressType, ifaceNames []string) (string, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return "", fmt.Errorf("get interfaces: %w", err)
	}

	// Candidate addresses.
	var candidates []string

	for _, iface := range ifaces {
		if len(ifaceNames) > 0 && contains(ifaceNames, iface.Name) {
			// Filter out interfaces if name doesn't match.
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			// Skip interface if we can't get addresses.
			continue
		}

		for _, addr := range addrs {
			ip, _, err := net.ParseCIDR(addr.String())

			if err != nil {
				panic("Can't parse interface address")
			}

			if GetAddressType(ip) != addrType {
				// Skip if address type doesn't match.
				continue
			}

			if isPrivate(ip) {
				// Skip if address is private.
				continue
			}

			candidates = append(candidates, ip.String())
		}
	}

	// Sort to get predictive results.
	sort.Strings(candidates)

	if len(candidates) == 0 {
		return "", nil
	}

	return candidates[0], nil
}
