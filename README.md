cf-ddns-go
===
wacky'6 second iteration for CloudFlare DDNS.

Key features:
* IPv6 DDNS support
* Network interface address detection
* Daemon (as a service) or One-Shot (as a command line tool)
* [ More to come ]


## Usage

```shell
CF_TOKEN=abcdef cf-ddns -6 hostname.example.com -D
```

* `CF_TOKEN` is the CloudFlare API Token that grants access to:
  * Zone/Zone: READ
  * Zone/DNS:  READ & WRITE
* `-6 hostname.example.com` instructs the program to set DNS AAAA record for hostname.example.com
* `-D` instructs the progrma to run like a command line tool, instead of going into daemon mode
* By default, the program detects IPv6 address by looking at the network interface IP addresses


## Daemon Installation (Debian / Ubuntu)

1. Download (or build) the binary, put it to `/usr/local/bin/cf-ddns`
2. Add executable permission `chmod +x /usr/local/bin/cf-ddns`
3. Modify and add the sample systemd unit file to OS
    * Sample file: `etc/systemd/system/cf-ddns.service`
    * Set `CF_TOKEN` environment variable to your Cloudflare API token
    * Set desired hostname (`V6_HOST`)
4. Enable the cf-ddns service in systemd
