# SLES Docker Setup Tool

Enterprise Docker installation and configuration tool, designed specifically for SLES 15+. Zero dependencies, one-click operation.

## Features

- ✅ **Truly zero dependencies** - Go static compilation, single binary, no Python or other runtime required
- ✅ **Interactive configuration** - Friendly menu guidance, easy for beginners to configure correctly
- ✅ **Enterprise best practices** - Built-in Registry, HTTP Proxy, and container network CIDR configuration
- ✅ **Network conflict detection** - Automatically detects if CIDR conflicts with internal network
- ✅ **Idempotent** - Safe to run multiple times, automatically backs up configuration files
- ✅ **Skip mechanism** - Each configuration can be skipped temporarily for manual setup later

## Usage

### One-click Run

```bash
curl -sSL https://internal.example.com/tools/setup-docker | sudo bash
```

### Or Download and Run

```bash
curl -sSL https://internal.example.com/tools/setup-docker -o setup-docker
chmod +x setup-docker
sudo ./setup-docker
```

## Configuration Options

| Configuration | Description | Default Value |
|---------------|-------------|---------------|
| **Registry** | Enterprise internal mirror registry address | registry.example.com |
| **HTTP Proxy** | Enterprise HTTP/HTTPS proxy | http://proxy.example.com:8080 |
| **Container CIDR** | Container bridge network, avoid conflicts with internal network | 172.31.0.0/16 |

## Configuration File Locations

- Docker daemon config: `/etc/docker/daemon.json`
- Systemd proxy config: `/etc/systemd/system/docker.service.d/http-proxy.conf`

## Development

### Requirements

- Go 1.21+

### Build

```bash
# Build
make build

# Build and compress
make compress

# Run tests
make test
```

### Enterprise Customization

Modify the default value constants in `internal/config/config.go` to adapt to your enterprise environment:

```go
const (
    DefaultRegistry   = "your-registry.com"
    DefaultHTTPProxy  = "http://your-proxy:8080"
    DefaultHTTPSProxy = "http://your-proxy:8080"
    DefaultNoProxy    = "localhost,127.0.0.1,.your-company.com"
    DefaultBIP        = "172.31.0.1/16"
)
```

To modify the Docker package name, update the `installDockerPackages` function in `internal/install/install.go`.

## Project Structure

```
sles-docker-setup/
├── cmd/
│   └── main.go          # Program entry, workflow orchestration
├── internal/
│   ├── install/         # Docker installation logic
│   ├── config/          # Configuration handling (Registry/Proxy/CIDR)
│   ├── system/          # System checks, service management, utilities
│   └── ui/              # Interactive UI, color output, forms
├── go.mod
├── Makefile
└── README.md
```

## Important Notes

- Must run with sudo
- SLES 15+ only
- Docker service will restart automatically after configuration changes
- Current user is automatically added to the `docker` group - requires re-login or running `newgrp docker` to take effect
