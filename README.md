cf-ddns-go
===
wacky'6 second iteration for CloudFlare DDNS.

Key features:
* IPv6 DDNS support
* Network interface address detection
* Daemon (as a service) or One-Shot (as a command line tool)
* [ More to come ]

## Installation (Debian / Ubuntu)

1. Download (or build) the binary, put it to `/usr/local/bin/cf-ddns`
2. Add executable permission `chmod +x /usr/local/bin/cf-ddns`
3. Modify and add the sample systemd unit file to OS
    * Sample file: `etc/systemd/system/cf-ddns.service`
    * Set `CF_KEY` and `CF_EMAIL` environment variable to your Cloudflare access key and email address
    * Set desired hostname (`V6_HOST`)
4. Enable the cf-ddns service in systemd
