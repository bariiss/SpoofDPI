# SpoofDPI 🛡️

A simple, fast, and cross-platform anti-censorship proxy designed to bypass **Deep Packet Inspection (DPI)**. SpoofDPI works by fragmenting TLS Client Hello packets and providing flexible DNS and proxy options to evade censorship systems.

![SpoofDPI Banner](https://user-images.githubusercontent.com/45588457/148035986-8b0076cc-fefb-48a1-9939-a8d9ab1d6322.png)

# Usage 🚀
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

## Features ✨
- **Bypass DPI**: Fragments TLS Client Hello to evade DPI-based censorship.
- **System Proxy Integration**: Automatically sets system-wide proxy on macOS (and optionally on Linux).
- **Flexible DNS**: Supports system DNS, custom DNS, and DNS-over-HTTPS (DoH).
- **Pattern-based Whitelisting**: Only bypass DPI for domains matching user-defined regex patterns.
- **IPv4/IPv6 Support**: Optionally restrict DNS to IPv4 only.
- **Configurable Timeout & Window Size**: Fine-tune fragmentation and connection behavior.
- **Silent & Debug Modes**: Control output verbosity.
- **Docker Support**: Run easily in containers.

---

## Installation 📦

### Using Makefile (macOS - Recommended) 🍎
The easiest way to install and manage SpoofDPI on macOS is using the included Makefile, which provides automated service management with launchd:

```bash
# Clone the repository
git clone https://github.com/bariiss/SpoofDPI.git
cd SpoofDPI

# Build, install, and start as a service (all in one)
make all

# Or step by step:
make build           # Build the binary
make install         # Install to ~/go/bin
make service-install # Create launchd service
make service-start   # Start the service
```

#### Service Management Commands
```bash
make service-start      # Start the service
make service-stop       # Stop the service
make service-restart    # Restart the service
make service-status     # Check service status
make service-logs       # View service logs
make service-uninstall  # Remove service completely
```

#### Custom Configuration
You can configure the service with custom parameters:
```bash
# Configure with custom settings
make service-config PORT=8080 ENABLE_DOH=false SYSTEM_PROXY=true

# Available parameters:
# PORT=8080                    - Proxy port
# DNS=1.1.1.1                 - DNS server
# ADDR=0.0.0.0                - Bind address  
# WINDOW_SIZE=1               - TLS fragmentation window size
# ENABLE_DOH=true             - Enable DNS over HTTPS
# SYSTEM_PROXY=false          - Enable system-wide proxy
```

The Makefile automatically:
- Detects your username and creates a personalized service (`com.<username>.spoofdpi`)
- Manages launchd service lifecycle
- Handles service logs (`/tmp/spoofdpi.log` and `/tmp/spoofdpi.err`)
- Provides easy configuration updates

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

### Docker 🐳
```bash
# Using Makefile
make docker-build    # Build Docker image
make docker-run      # Run Docker container

# Manual Docker usage
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

## Quick Start 🚀

### macOS with Makefile (Recommended)
```bash
# Clone and setup as a service
git clone https://github.com/bariiss/SpoofDPI.git
cd SpoofDPI
make all

# Check service status
make service-status

# View logs
make service-logs
```

### macOS (Manual)
Just run the `spoofdpi` command. The proxy will be set up automatically.

### Linux
Start `spoofdpi` and launch your browser with the following command:
```bash
google-chrome --proxy-server="http://127.0.0.1:8080"
```

---

## Command Line Options ⚙️
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

## How It Works 🔍
- **HTTP**: Serves as a proxy for HTTP requests (no DPI bypass, as most censorship targets HTTPS).
- **HTTPS**: Fragments the TLS Client Hello packet (either in two parts or user-defined window size) to evade DPI systems that inspect only the first chunk.
- **DNS**: Supports system DNS, custom DNS, and DNS-over-HTTPS for flexible name resolution.
- **Pattern Matching**: DPI bypass is only applied to domains matching the provided regex patterns (if any).

---

## Configuration & Advanced Usage 🛠️

### Service Management (macOS)
When using the Makefile approach, SpoofDPI runs as a persistent launchd service:

```bash
# View all available commands
make help

# Show current configuration
make show-config

# Configure service with custom parameters
make service-config PORT=9090 DNS=1.1.1.1 ENABLE_DOH=true

# Apply configuration changes
make service-restart

# Monitor service
make service-status
make service-logs

# Development mode (run without service)
make dev-run
```

The service automatically:
- Starts on system boot (`RunAtLoad=true`)
- Restarts if it crashes (`KeepAlive=true`)
- Logs to `/tmp/spoofdpi.log` and `/tmp/spoofdpi.err`
- Uses your username in the service name (`com.<username>.spoofdpi`)

### Manual Configuration
- **System Proxy**: On macOS, system proxy is set automatically (may require admin privileges). On Linux, set your browser's proxy manually.
- **Allowed Patterns**: Use `-pattern` multiple times to specify regexes for domains to bypass DPI.
- **Window Size**: Use `-window-size` to control TLS fragmentation granularity.
- **Debugging**: Use `-debug` for verbose logs.
- **Silent Mode**: Use `-silent` to suppress banner and info output.

---

## Troubleshooting 🔧

### macOS Service Issues
```bash
# Check if service is running
make service-status

# View service logs
make service-logs

# Restart service if having issues
make service-restart

# Completely reinstall service
make service-uninstall
make service-install
make service-start

# Check launchd directly
launchctl print gui/$(id -u)/com.$(whoami).spoofdpi
```

### Common Issues
- **Permission denied**: Make sure `~/go/bin` is in your PATH and the binary has execute permissions
- **Service won't start**: Check logs with `make service-logs` and ensure no other process is using the configured port
- **DNS issues**: Try different DNS servers with `make service-config DNS=1.1.1.1`
- **Connection problems**: Adjust window size with `make service-config WINDOW_SIZE=2`

---

## Docker Compose Example 🐳
See [`docker-compose.yml`](./docker-compose.yml) for a ready-to-use configuration.

---

## Project Structure 📁
- `Makefile`      : macOS service management and build automation
- `cmd/spoofdpi/` : Main entrypoint
- `proxy/`        : Proxy server logic (HTTP/HTTPS, handlers)
- `dns/`          : DNS resolver logic (system, custom, DoH)
- `packet/`       : HTTP/TLS packet parsing and manipulation
- `util/`         : Utilities (args, config, logging, OS integration)
- `version/`      : Versioning
- `_docs/`        : Additional documentation

---

## Technical Details 🔧
- **Go Version**: Built with Go 1.24.2
- **License**: Apache License 2.0
- **Dependencies**:
  - github.com/miekg/dns v1.1.65
  - github.com/pterm/pterm v0.12.80
  - github.com/rs/zerolog v1.34.0

---

## License 📄
This project is licensed under the Apache License 2.0. See [LICENSE](./LICENSE) for details.

---

## Inspirations 💡
- [Green Tunnel](https://github.com/SadeghHayeri/GreenTunnel) by @SadeghHayeri
- [GoodbyeDPI](https://github.com/ValdikSS/GoodbyeDPI) by @ValdikSS

---

## Contributing 🤝
Pull requests and issues are welcome! Please see the code and documentation for contribution guidelines.
