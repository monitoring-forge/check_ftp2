# check_ftp2

Nagios `check_ftp` plugin alternative supporting TLSv1.2 and explicit/implicit TLS mode.
Not implemented full feature, only we need.

## Installation

### From GitHub Releases

Pre-built binaries for Linux are available from the [GitHub Releases](https://github.com/monitoring-forge/check_ftp2/releases) page.

For Linux (amd64):

```bash
wget https://github.com/monitoring-forge/check_ftp2/releases/latest/download/check_ftp2_linux_amd64.tar.gz
tar -xzf check_ftp2_linux_amd64.tar.gz
chmod +x check_ftp2
sudo mv check_ftp2 /usr/local/bin/
check_ftp2 --version
```

> Replace `linux_amd64` with the archive name that matches your architecture.

### From source

Requires Go 1.25 or later.

```bash
git clone https://github.com/monitoring-forge/check_ftp2.git
cd check_ftp2
make
sudo cp check_ftp2 /usr/local/bin/
```

## Usage

```
Usage:
  check_ftp2 [OPTIONS]

Application Options:
      --timeout=    Timeout to wait for connection (default: 10s)
  -H, --hostname=   IP address or Host name (default: 127.0.0.1)
  -p, --port=       Port number (default: 21)
  -S, --ssl         use TLS
      --sni=        specify hostname for SNI
      --explicit    Use Explicit TLS mode
  -4                use tcp4 only
  -6                use tcp6 only
      --verify-ssl  Verify SSL certificate, --sni must be specified
  -v, --version     Show version

Help Options:
  -h, --help        Show this help message
```

## Options

| Option | Short | Description |
|--------|-------|-------------|
| `--timeout` | - | Connection timeout. Accepts values like `10s`, `1m`. Default is `10s`. |
| `--hostname` | `-H` | Target FTP server IP address or host name. Default is `127.0.0.1`. |
| `--port` | `-p` | Target FTP server port. Default is `21`. |
| `--ssl` | `-S` | Connect with implicit TLS (FTPS). The TLS handshake is performed immediately after TCP connection. |
| `--explicit` | - | Connect with explicit TLS (FTPES). The connection starts as plain FTP and upgrades to TLS via `AUTH TLS`. |
| `--sni` | - | Server Name Indication (SNI) hostname sent during the TLS handshake. Required when `--verify-ssl` is used. |
| `--verify-ssl` | - | Verify the server TLS certificate. `--sni` must also be specified. |
| `-4` | - | Use IPv4 only. |
| `-6` | - | Use IPv6 only. |
| `--version` | `-v` | Show the version and exit. |
| `--help` | `-h` | Show help and exit. |

### Connection mode matrix

| Mode | Options | Description |
|------|---------|-------------|
| Plain FTP | (none) | Unencrypted FTP connection. |
| Implicit TLS | `-S` / `--ssl` | TLS handshake before FTP protocol. Commonly uses port `990`. |
| Explicit TLS | `--explicit` | Plain FTP first, then `AUTH TLS`. Commonly uses port `21`. |

## Examples

### Plain FTP

```bash
check_ftp2 -H ftp.example.com -p 21
```

### Implicit TLS (FTPS)

```bash
check_ftp2 -H ftp.example.com -p 990 -S
```

### Explicit TLS without certificate verification

Useful for self-signed certificates or when the server hostname does not match the certificate.

```bash
check_ftp2 -H ftp.example.com -p 21 --explicit
```

### Explicit TLS with certificate verification

```bash
check_ftp2 -H ftp.example.com -p 21 --explicit --verify-ssl --sni ftp.example.com
```

### IPv4 only

```bash
check_ftp2 -H ftp.example.com -p 21 -4
```

## Output

### OK

```
FTP OK - 0.003 second response time on ftp.example.com port 21 [220 Welcome...]|time=0.003401s;;;0.000000;10.000000
```

* Exit code: `0`
* `time` is the measured response time in seconds.

### CRITICAL

```
FTP CRITICAL: connection failed: EOF on ftp.example.com port 21 []
```

* Exit code: `2`
* The server closed the connection unexpectedly or denied access.

```
FTP CRITICAL: connection failed: dial tcp 127.0.0.1:21: connect: connection refused on 127.0.0.1 port 21 []
```

* Exit code: `2`
* The target host/port is unreachable or no FTP server is listening.

```
FTP CRITICAL: connection failed: tls: failed to verify certificate: x509: certificate is not valid for any names, but wanted to match ftp.example.com on ftp.example.com port 21 [220-...\nAUTH TLS\n234 AUTH TLS OK.]
```

* Exit code: `2`
* Certificate verification failed. Use `--sni` with the correct hostname or omit `--verify-ssl`.

### UNKNOWN

```
FTP UNKNOWN: verify-ssl is specified but sni is not specified
```

* Exit code: `3`
* Invalid option combination. `--verify-ssl` requires `--sni`.

```
FTP UNKNOWN: both tcp4 and tcp6 are specified
```

* Exit code: `3`
* `-4` and `-6` cannot be used together.

## Development

### Run unit tests

```bash
make check
```

### Run integration tests

Integration tests start real FTP servers with Docker Compose.

```bash
make integration-test
```

This requires Docker, Docker Compose, `openssl`, `nc`, and `go`.

## License

[MIT License](LICENSE)

Copyright (c) 2020 Masahiro Nagano

