# go-mal-pkgs

A security tool to detect malicious Go packages by verifying checksums in go.sum against the original source code, bypassing the Go Module Proxy cache.

## Overview

This tool helps protect against supply chain attacks that exploit the Go Module Proxy cache, as recently demonstrated by the malicious `github.com/boltdb-go/bolt` package.
These attacks take advantage of the Go Module Proxy's indefinite caching mechanism to serve malicious code even after the source repository has been cleaned.

For more details about this type of attack, see:
- [Socket - Go Supply Chain Attack: Malicious Package Exploits Go Module Proxy Caching for Persistence](https://socket.dev/blog/malicious-package-exploits-go-module-proxy-caching-for-persistence)

The tool works by:
1. Reading the go.sum file from your Go project
2. Fetching the original source code directly from the repositories (bypassing the Go Module Proxy)
3. Calculating checksums using Go's official algorithm
4. Comparing the checksums to detect any mismatches that could indicate compromised packages

## Installation

```bash
go install github.com/knqyf263/go-mal-pkgs@latest
```

## Usage

Basic usage:
```bash
go-mal-pkgs /path/to/your/go/project
```

The tool will analyze your project's go.sum file and verify each module's integrity.

### Example Output

Successful verification:
```bash
$ go-mal-pkgs /path/to/project
Checking golang.org/x/text@v0.3.0...
✅ Verified golang.org/x/text@v0.3.0
Checking github.com/stretchr/testify@v1.8.4...
✅ Verified github.com/stretchr/testify@v1.8.4
```

Detection of a potentially malicious package:
```bash
$ go-mal-pkgs /path/to/compromised-project
Checking github.com/boltdb-go/bolt@v1.3.1...
⚠️  WARNING: checksum mismatch for github.com/boltdb-go/bolt@v1.3.1:
expected: h1:abc123def456...
actual: h1:xyz789uvw012...
```

## How it Works

The tool implements Go's checksum verification algorithm using the `golang.org/x/mod/sumdb/dirhash` package. When it detects a mismatch between the checksum in your go.sum and the one calculated from the source repository, it could indicate:

1. A compromised module in the Go Module Proxy cache
2. A malicious package that has been tampered with
3. A typosquatting attack (like the boltdb-go/bolt case)
4. Other supply chain security issues

## Limitations

- The tool currently only supports modules hosted on GitHub
- go.mod checksums are not verified at this time
- Network connectivity is required to fetch source code

---

## Note
This is an experimental tool created as a proof of concept in response to the recent Go module proxy cache exploitation. The entire codebase was generated using Claude 3.5 Sonnet AI assistant.