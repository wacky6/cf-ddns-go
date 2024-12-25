package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/jessevdk/go-flags"
	"github.com/miekg/dns"

	"github.com/wacky6/cf-ddns-go/ddns"
	"github.com/wacky6/cf-ddns-go/util"
)

var opts struct {
	// ** IPv4 not implemented **
	// Addr4     string        `short:"4" long:"addr4" description:"Fully qualified domain name to set IPv4 A record" value-name:"<fqdn>"`
	// Mode4     string        `          long:"mode4" description:"IP detection mode for IPv4" choice:"iface" choice:"probe" default:"probe"`
	// Iface4    []string      `          long:"iface4" description:"Interfaces to check for IPv4" default:"" value-name:"<interface_name>" default-mask:"all-interfaces"`
	// Interval4 time.Duration `          long:"interval4" description:"Number of seconds between consecutive IPv4 address checks" value-name:"<duration>" default:"60s"`
	Addr6     string        `short:"6" long:"addr6" description:"Fully qualified domain name to set IPv6 AAAA record" value-name:"<fqdn>"`
	Mode6     string        `          long:"mode6" description:"IP detection mode for IPv6" choice:"iface" default:"iface"`
	Iface6    []string      `          long:"iface6" description:"Interfaces to check for IPv6" default:"" value-name:"<interface_name>" default-mask:"all-interfaces"`
	Interval6 time.Duration `          long:"interval6" description:"Time between consecutive IPv4 address checks" value-name:"<duration>" default:"20s"`

	DNSServer []string `short:"r" long:"resolver" description:"DNS resolvers to check for existing records" default:"default" value-name:"<dns_resolver>" default-mask:"CF & Google DNS"`
	OneShot   bool     `short:"D" long:"one-shot" description:"Detect and set DNS record once, don't enter daemon mode"`
	Version   bool     `short:"V" long:"version" description:"Print version information and exit"`
}

// Build time constants, should be provided by linker.
var buildTime string   // The format should be strptime "%Y-%m-%d %H:%M:%S UTC".
var buildCommit string // Git commit SHA
var buildTag string    // Git version tag (if available)

func main() {
	_, err := flags.Parse(&opts)
	if err != nil {
		os.Exit(1)
	}

	if opts.Version {
		printVersionInfoOneLine(os.Stderr)
		os.Exit(0)
	}

	if len(opts.DNSServer) == 1 && opts.DNSServer[0] == "default" {
		opts.DNSServer = []string{
			// Cloudflare DNS
			"1.1.1.1",
			"1.0.0.1",
			"2606:4700:4700::1111",
			"2606:4700:4700::1001",

			// Google public DNS
			"8.8.8.8",
			"8.8.4.4",
			"2001:4860:4860::8888",
			"2001:4860:4860::8844",
		}
	}

	if !opts.OneShot {
		printVersionInfoOneLine(os.Stderr)
	}

	provider, err := ddns.CreateCloudFlareProvider()
	if err != nil {
		fmt.Fprintf(os.Stderr, "can't initialize ddns provider: %v\n", err)
		os.Exit(1)
	}

	err = provider.VerifyConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "can't verify ddns provider config: %v\n", err)
		os.Exit(1)
	}

	if len(opts.Addr6) > 0 {
		detectFn := func() string {
			result, err := util.DetectAddress(opts.Mode6, util.IP6, opts.Iface6)
			if err != nil {
				return ""
			}
			return result
		}
		resolveFn := getResolveFn(opts.DNSServer, util.IP6, opts.Addr6)
		updateFn := func(addr string) error {
			return provider.SetRecord(opts.Addr6, ddns.Record{Type: "AAAA", Content: addr})
		}

		var updateInterval time.Duration
		if opts.OneShot {
			updateInterval = 0 * time.Second
		} else {
			updateInterval = opts.Interval6
		}

		result, err := ddnsLoop("ipv6", resolveFn, detectFn, updateFn, updateInterval)
		if opts.OneShot {
			if err != nil {
				fmt.Fprintf(os.Stdout, "FAIL ipv6 <unknown>\n")
			} else {
				fmt.Fprintf(os.Stdout, "OK ipv6 %v\n", result)
			}
		}
	}
}

// fillEmpty returns "<empty>" if the provided string is empty
func fillEmpty(str string) string {
	if len(str) == 0 {
		return "<empty>"
	}

	return str
}

// getResolveFn returns a DNS resolve function. If dnsServerSpec is non-empty, it queries the provied servers.
// Otherwise it returns a function to use operating system's DNS resolver.
func getResolveFn(dnsServerSpec []string, addrType util.AddressType, fqdn string) func() string {
	var dnsRecordType uint16
	switch addrType {
	case util.IP4:
		dnsRecordType = dns.TypeA
	case util.IP6:
		dnsRecordType = dns.TypeAAAA
	default:
		log.Fatalf("invalid address type: %v\n", addrType)
		return func() string { return "" }
	}

	// Use provided DNS server.
	if len(dnsServerSpec) > 0 {
		return func() string { return util.Resolve(dnsServerSpec, dnsRecordType, fqdn) }
	}

	// Use OS DNS resolver.
	return func() string { return util.ResolveOs(dnsRecordType, fqdn) }
}

// ddnsLoop executes the DDNS resolve-detect-update loop. Returns the updated ip address.
func ddnsLoop(
	logPrefix string,
	resolveFn func() string, // resolveFn returns the resolved DNS record for `fqdn`
	detectFn func() string, // detectFn returns the detected ip address
	updateFn func(addr string) error, // updateFn upadtes the record to addr, and returns the error
	interval time.Duration, // interval between two consecutive checks, `0` means running as one-shot
) (string, error) {
	var oneShot bool = interval == 0*time.Second // Whether running as a one-shot function call.

	const DNSQueryTTL = 10 * time.Minute // The amount of time to wait, before querying a real DNS server after a valid response
	var lastDNSQueryTime time.Time       // Last time that we query a DNS server
	var lastResolvedAddress string       // Last address that is resolved by DNS server, or sent to DNS provider

	const statusTTL = 1 * time.Hour // The amount of time to wait, before refreshing the status message (log)
	var lastStatusTime time.Time    // Last time that we shown a status message

	// A helper function to display status messages.
	printStatusF := func(formatStr string, args ...interface{}) {
		if lastStatusTime.Add(statusTTL).Before(time.Now()) {
			fmt.Fprintf(os.Stderr, formatStr, args...)
			lastStatusTime = time.Now()
		}
	}

	for {
		chanDetect := make(chan string)
		chanResolve := make(chan string)

		go func() { chanDetect <- detectFn() }()
		go func() {
			if lastDNSQueryTime.Add(DNSQueryTTL).Before(time.Now()) {
				// DNS query is stale, resolve the address.
				lastResolvedAddress = resolveFn()
				lastDNSQueryTime = time.Now()
			}

			chanResolve <- lastResolvedAddress
		}()

		detected := <-chanDetect
		resolved := <-chanResolve

		if len(detected) == 0 {
			if oneShot {
				return "", fmt.Errorf("failed to detect ip address")
			}
			printStatusF("%v failed to detect address\n", logPrefix)
			time.Sleep(interval)
			continue
		}

		if resolved != detected {
			err := updateFn(detected)
			if err != nil {
				if oneShot {
					return "", fmt.Errorf("failed to update record: %w", err)
				}
				printStatusF("%v failed to update record: %v\n", logPrefix, err.Error())
				time.Sleep(interval)
				continue
			}
		}

		if oneShot {
			return detected, nil
		}

		if resolved != detected {
			fmt.Fprintf(os.Stderr, "%v updated %v\n", logPrefix, detected)
			lastStatusTime = time.Now()

			// We just set the DNS record, pretend it to be the response of the next DNS query.
			// This gives time for DNS records to propagate, and avoids querying DNS servers excessively.
			lastDNSQueryTime = time.Now()
			lastResolvedAddress = detected
		} else {
			printStatusF("%v up-to-date %v\n", logPrefix, detected)
		}

		time.Sleep(interval)
	}
}

func printVersionInfoOneLine(out io.Writer) {
	var visibleTag string
	if len(buildTag) > 0 {
		visibleTag = buildTag
	} else {
		visibleTag = "tip_of_tree"
	}

	var visibleCommitSHA string
	if len(buildCommit) > 0 {
		visibleCommitSHA = buildCommit
	} else {
		visibleCommitSHA = "HEAD"
	}

	if len(buildTime) > 0 {
		fmt.Fprintf(os.Stderr, "cf-ddns version %v %v build_at %v\n", visibleTag, visibleCommitSHA, buildTime)
	} else {
		fmt.Fprintf(os.Stderr, "cf-ddns version %v %v\n", visibleTag, visibleCommitSHA)
	}
}
