package main

import (
	"fmt"
	"os"
	"time"

	"github.com/jessevdk/go-flags"

	"../ddns"
	"../util"
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

	OneShot bool `short:"D" long:"one-shot" description:"Detect and set DNS record once, don't enter daemon mode"`
}

func main() {
	_, err := flags.Parse(&opts)
	if err != nil {
		os.Exit(1)
	}

	if len(opts.Addr6) > 0 {
		err = run(func(provider ddns.Provider, lastAddr string) (string, error) {
			return runIPv6(provider, opts.Addr6, opts.Mode6, opts.Iface6, lastAddr)
		}, opts.Interval6)
	}
}

func daemonize(fn func() error, interval time.Duration) {
	for {
		fn()
		time.Sleep(interval)
	}
}

func run(
	runFn func(provider ddns.Provider, lastAddr string) (string, error),
	interval time.Duration) error {

	provider, err := ddns.CreateCloudFlareProvider()
	if err != nil {
		fmt.Fprintf(os.Stderr, "can't initialize ddns provider: %v\n", err)
		return err
	}

	if opts.OneShot {
		addr, err := runFn(provider, "")
		if err != nil {
			fmt.Fprintf(os.Stdout, "FAIL\n")
			fmt.Fprintf(os.Stderr, "failed to update ddns: %v\n", err)
			return err
		}

		fmt.Fprintf(os.Stdout, "OK %v\n", addr)
		return nil
	}

	// Run as daemon.
	err = provider.VerifyConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "can't start as daemon: %v\n", err)
		return err
	}

	var lastAddr string
	daemonize(func() error {
		addr, err := runFn(provider, lastAddr)
		lastAddr = addr
		return err
	}, interval)
	return nil
}

// runIPv6 returns the ip address set to the DNS record.
func runIPv6(provider ddns.Provider, fqdn string, mode string, ifaces []string, currentAddr string) (string, error) {
	addr, err := util.DetectAddress(mode, util.IP6, ifaces)
	if err != nil {
		return "", err
	}

	if len(addr) == 0 {
		return "", fmt.Errorf("no public ipv6 address found")
	}

	if currentAddr == addr {
		return addr, nil
	}

	err = provider.SetRecord(fqdn, ddns.Record{Type: "AAAA", Content: addr})
	if err != nil {
		return "", err
	}

	return addr, nil
}
