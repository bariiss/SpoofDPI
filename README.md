# SpoofDPI

A simple, fast, and cross-platform anti-censorship proxy designed to bypass **Deep Packet Inspection (DPI)**. SpoofDPI works by fragmenting TLS Client Hello packets and providing flexible DNS and proxy options to evade censorship systems.

![SpoofDPI Banner](https://user-images.githubusercontent.com/45588457/148035986-8b0076cc-fefb-48a1-9939-a8d9ab1d6322.png)

# Usage
```
Usage: spoofdpi [options...]
  -addr string
        listen address (default "127.0.0.1")
  -debug
        enable debug output
  -dns-addr string
        dns address (default "8.8.8.8")
  -dns-ipv4-only
        resolve only version 4 addresses
  -dns-port value
        port number for dns (default 53)
  -enable-doh
        enable 'dns-over-https'
  -pattern value
        bypass DPI only on packets matching this regex pattern; can be given multiple times
  -port value
        port (default 8080)
  -silent
        do not show the banner and server information at start up
  -system-proxy
        enable system-wide proxy (default true)
  -timeout value
        timeout in milliseconds; no timeout when not given
  -v    print spoofdpi's version; this may contain some other relevant information
  -window-size value
        chunk size, in number of bytes, for fragmented client hello,
        try lower values if the default value doesn't bypass the DPI;
        when not given, the client hello packet will be sent in two parts:
        fragmentation for the first data packet and the rest
```

---

## Features
- **Bypass DPI**: Fragments TLS Client Hello to evade DPI-based censorship.
- **System Proxy Integration**: Automatically sets system-wide proxy on macOS (and optionally on Linux).
- **Flexible DNS**: Supports system DNS, custom DNS, and DNS-over-HTTPS (DoH).
- **Pattern-based Whitelisting**: Only bypass DPI for domains matching user-defined regex patterns.
- **IPv4/IPv6 Support**: Optionally restrict DNS to IPv4 only.
- **Configurable Timeout & Window Size**: Fine-tune fragmentation and connection behavior.
- **Silent & Debug Modes**: Control output verbosity.
- **Docker Support**: Run easily in containers.

---

## Installation

### Pre-built Binary
A detailed installation guide is available in [`_docs/INSTALL.md`](./_docs/INSTALL.md).

Quick install (macOS/Linux):
```bash
curl -fsSL https://raw.githubusercontent.com/bariiss/SpoofDPI/main/install.sh | bash -s <platform>
```
Replace `<platform>` with one of: `darwin-amd64`, `darwin-arm64`, `linux-amd64`, `linux-arm`, `linux-arm64`, `linux-mips`, `linux-mipsle`.

### Go
```bash
go install github.com/bariiss/SpoofDPI/cmd/spoofdpi@latest
```

### Docker
```bash
docker run --rm -it \
  -e WINDOW_SIZE=1 \
  -e APP_PORT=8080 \
  -e APP_ADDR=0.0.0.0 \
  -e DOH_ENABLED=false \
  -e DNS_ADDR=8.8.8.8 \
  -e DNS_PORT=53 \
  -e SYSTEM_PROXY=false \
  -e DEBUG_MODE=true \
  -p 8080:8080 \
  ghcr.io/bariiss/spoofdpi:latest
```
A sample `docker-compose.yml` is provided in the repository.

---

## Quick Start

### macOS
Sadece `spoofdpi` komutunu çalıştırın. Proxy otomatik olarak ayarlanır.

### Linux
`spoofdpi`'yi başlatın ve tarayıcınızı aşağıdaki gibi başlatın:
```bash
google-chrome --proxy-server="http://127.0.0.1:8080"
```

---

## Command Line Options
```
Usage: spoofdpi [options...]
  -addr string           listen address (default "127.0.0.1")
  -port value            port (default 8080)
  -dns-addr string       dns address (default "8.8.8.8")
  -dns-port value        port number for dns (default 53)
  -dns-ipv4-only         resolve only version 4 addresses
  -enable-doh            enable 'dns-over-https'
  -pattern value         bypass DPI only on packets matching this regex pattern; can be given multiple times
  -window-size value     chunk size, in number of bytes, for fragmented client hello
  -timeout value         timeout in milliseconds; no timeout when not given
  -system-proxy          enable system-wide proxy (default true)
  -debug                 enable debug output
  -silent                do not show the banner and server information at start up
  -v                     print spoofdpi's version and exit
```

---

## How It Works
- **HTTP**: Serves as a proxy for HTTP requests (no DPI bypass, as most censorship targets HTTPS).
- **HTTPS**: Fragments the TLS Client Hello packet (either in two parts or user-defined window size) to evade DPI systems that inspect only the first chunk.
- **DNS**: Supports system DNS, custom DNS, and DNS-over-HTTPS for flexible name resolution.
- **Pattern Matching**: DPI bypass is only applied to domains matching the provided regex patterns (if any).

---

## Configuration & Advanced Usage
- **System Proxy**: On macOS, system proxy is set automatically (may require admin privileges). On Linux, set your browser's proxy manually.
- **Allowed Patterns**: Use `-pattern` multiple times to specify regexes for domains to bypass DPI.
- **Window Size**: Use `-window-size` to control TLS fragmentation granularity.
- **Debugging**: Use `-debug` for verbose logs.
- **Silent Mode**: Use `-silent` to suppress banner and info output.

---

## Docker Compose Example
See [`docker-compose.yml`](./docker-compose.yml) for a ready-to-use configuration.

---

## Project Structure
- `cmd/spoofdpi/` : Main entrypoint
- `proxy/`        : Proxy server logic (HTTP/HTTPS, handlers)
- `dns/`          : DNS resolver logic (system, custom, DoH)
- `packet/`       : HTTP/TLS packet parsing and manipulation
- `util/`         : Utilities (args, config, logging, OS integration)
- `version/`      : Versioning
- `_docs/`        : Additional documentation

---

## License
This project is licensed under the Apache License 2.0. See [LICENSE](./LICENSE) for details.

---

## Inspirations
- [Green Tunnel](https://github.com/SadeghHayeri/GreenTunnel) by @SadeghHayeri
- [GoodbyeDPI](https://github.com/ValdikSS/GoodbyeDPI) by @ValdikSS

---

## Contributing
Pull requests and issues are welcome! Please see the code and documentation for contribution guidelines.
