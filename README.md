# redis-Tui

A Redis Text-based UI client in CLI.

*Forked from https://github.com/mylxsw/redis-tui*

![](./preview.gif)

## Features

- **REDIS_URL Support**: Connect using standard Redis URLs from environment variable or command line
- **TLS Support**: Secure connections with full certificate configuration
- **Command auto-completion**: Quick command entry with history
- **Multiple data type support**: String, List, Set, ZSet, Hash

## Installation

```bash
go install github.com/cwarden/redis-tui@latest
```

Or build from source:
```bash
make build
```

## Usage

### Basic Connection
```bash
# Connect to localhost
redis-tui

# Connect to specific host and port
redis-tui -h redis.example.com -p 6380

# With authentication
redis-tui -h redis.example.com -a mypassword
```

### Using REDIS_URL
```bash
# From environment variable
export REDIS_URL="redis://user:password@localhost:6379/0"
redis-tui

# From command line
redis-tui -url "redis://user:password@localhost:6379/0"

# TLS connection
redis-tui -url "rediss://user:password@secure.redis.com:6380/0"
```

### TLS/SSL Connections
```bash
# Enable TLS
redis-tui -h secure.redis.com -tls

# With client certificates
redis-tui -h secure.redis.com -tls \
  -tls-cert client.crt \
  -tls-key client.key \
  -tls-ca-cert ca.crt

# Skip certificate verification (not recommended for production)
redis-tui -h secure.redis.com -tls -tls-verify=false
```

### Cluster Mode
```bash
redis-tui -h cluster.redis.com -c
```

### Debug Mode
```bash
redis-tui -vvv
```

## Key Bindings

- `F1` / `Ctrl+N`: Command mode
- `F2` / `Ctrl+S`: Search keys
- `F3` / `Ctrl+K`: Key list
- `F4` / `Ctrl+F`: Focus on command input
- `F5` / `Ctrl+R`: Command result panel
- `F6` / `Ctrl+Y`: View list/set values
- `F7` / `Ctrl+A`: View string value
- `F9` / `Ctrl+O`: Output panel
- `Tab`: Switch focus between panels
- `j` / `k`: Navigate up and down within panel
- `Esc` / `Ctrl+Q`: Quit

## TODO

- [x] Solve the problem that the official environment generally disables the `KEYS` command to get the Key list
- [x] Command auto-completion function when executing commands
- [x] Command execution history function, you can quickly switch the previous command by pressing the up and down buttons
- [ ] The return value of the `SCAN` command is not formatted correctly


