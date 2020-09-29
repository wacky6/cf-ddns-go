cf-ddns-go
===
wacky'6 second iteration for CloudFlare DDNS.

## Installation

## Development
Development is based on code-server workflow. The built container image contains everything to reliably build the binary (and run tests).

1. Get source code `git clone https://github.com/wacky6/cf-ddns-go`
2. Build the development Docker image: `DOCKER_BUILDKIT=1 docker buid . -t cf-ddns-go-dev`
3. Run container: `docker run -itd -p <local_address>:<local_port>:9000 -v <git_checkout>:/cf-ddns-go cf-ddns-go-dev`
4. Run Chrome browser, navigate to `<local_address>:<local_port>`
