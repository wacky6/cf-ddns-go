package util

import (
	"net"

	"github.com/miekg/dns"
)

func lookupWithRetry(m *dns.Msg, server string, triesLeft uint) string {
	if triesLeft == 0 {
		return ""
	}

	in, err := dns.Exchange(m, net.JoinHostPort(server, "53"))

	if err != nil {
		return lookupWithRetry(m, server, triesLeft-1)
	}

	if in != nil && in.Rcode != dns.RcodeSuccess {
		return ""
	}

	for _, record := range in.Answer {
		return dns.Field(record, 1)
	}

	return ""
}

// Resolve resolves a DNS `recordType` dns.Type enum for `name` by querying servers.
func Resolve(servers []string, recordType uint16, name string) string {
	m := new(dns.Msg)
	m.Id = dns.Id()
	m.RecursionDesired = true
	m.Question = make([]dns.Question, 1)
	m.Question[0] = dns.Question{
		Name:   dns.Fqdn(name),
		Qtype:  uint16(recordType),
		Qclass: dns.ClassINET}

	ch := make(chan string, len(servers))

	for _, server := range servers {
		go func(ch chan<- string, server string) {
			ch <- lookupWithRetry(m, server, 3)
		}(ch, server)
	}

	for range servers {
		result := <-ch

		if len(result) > 0 {
			return result
		}
	}

	return ""
}

// ResolveOs returns a DNS `recordType` dns.Type enum for `name` by using operating system's resolver.
// Currently, only supports TypeA and TypeAAAA.
func ResolveOs(recordType uint16, name string) string {
	ips, err := net.LookupIP(name)

	if err != nil {
		return ""
	}

	for _, ip := range ips {
		if recordType == dns.TypeA && GetAddressType(ip) == IP4 {
			return ip.String()
		}
		if recordType == dns.TypeAAAA && GetAddressType(ip) == IP6 {
			return ip.String()
		}
	}

	return ""
}
