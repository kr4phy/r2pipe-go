# r2pipe-go

[![Go CI](https://github.com/kr4phy/r2pipe-go/actions/workflows/ci.yml/badge.svg)](https://github.com/kr4phy/r2pipe-go/actions/workflows/ci.yml)
[![GoDoc](https://pkg.go.dev/badge/github.com/kr4phy/r2pipe-go)](https://pkg.go.dev/github.com/kr4phy/r2pipe-go)
[![Go Report Card](https://goreportcard.com/badge/github.com/kr4phy/r2pipe-go)](https://goreportcard.com/report/github.com/kr4phy/r2pipe-go)

Go bindings for [radare2](https://github.com/radareorg/radare2), providing a convenient interface to execute r2 commands from Go programs.

## Features

- **Spawn Pipe Mode**: Communicate with radare2 through stdin/stdout pipes
- **Native Mode**: Load libr_core dynamically for better performance
- **API Mode**: Link directly against radare2 C libraries (requires build tag)
- **JSON Support**: Parse JSON output from radare2 commands
- **Event Handling**: Subscribe to stderr events

## Compatibility

| Component | Version |
|-----------|---------|
| Go | 1.23+ |
| radare2 | 5.x, 6.x |

## Installation

```bash
go get github.com/kr4phy/r2pipe-go
```

## Requirements

- **Spawn Pipe Mode**: radare2 must be installed and available in your PATH
- **Native Mode**: radare2 shared libraries (`libr_core.so`) must be installed
- **API Mode**: radare2 development headers and libraries required at compile time

## Quick Start

### Basic Example (Spawn Pipe)

The simplest way to use r2pipe-go is through the spawn pipe mode:

```go
package main

import (
    "fmt"

    r2pipe "github.com/kr4phy/r2pipe-go"
)

func main() {
    // Create a new pipe - this spawns a radare2 process
    r2p, err := r2pipe.NewPipe("malloc://256")
    if err != nil {
        panic(err)
    }
    defer r2p.Close()

    // Write data to memory
    _, err = r2p.Cmd("w Hello World")
    if err != nil {
        panic(err)
    }

    // Read back the string
    buf, err := r2p.Cmd("ps")
    if err != nil {
        panic(err)
    }
    fmt.Println(buf) // Output: Hello World
}
```

### Using Formatted Commands

```go
// Use Cmdf for formatted commands
result, err := r2p.Cmdf("px %d @ %d", 32, 0)
if err != nil {
    panic(err)
}
fmt.Println(result)
```

### Working with JSON Output

```go
// Parse JSON output into a struct
type FunctionInfo struct {
    Name   string `json:"name"`
    Offset uint64 `json:"offset"`
    Size   uint64 `json:"size"`
}

var functions []FunctionInfo
err := r2p.CmdjStruct("aflj", &functions)
if err != nil {
    panic(err)
}
for _, fn := range functions {
    fmt.Printf("Function: %s at 0x%x (size: %d)\n", fn.Name, fn.Offset, fn.Size)
}
```

### Native Mode (Better Performance)

For better performance, use the native mode which loads radare2 libraries dynamically:

```go
package main

import (
    "fmt"

    r2pipe "github.com/kr4phy/r2pipe-go"
)

func main() {
    // Create a native pipe - loads libr_core.so dynamically
    r2p, err := r2pipe.NewNativePipe("/bin/ls")
    if err != nil {
        panic(err)
    }
    defer r2p.Close()

    // Analyze and disassemble
    _, _ = r2p.Cmd("aaa")
    result, err := r2p.Cmd("pdf @ main")
    if err != nil {
        panic(err)
    }
    fmt.Println(result)
}
```

### API Mode (Direct C Linking)

For maximum performance, use the API mode which links directly against radare2 libraries. This requires the `r2api` build tag:

```go
//go:build r2api

package main

import (
    "fmt"

    r2pipe "github.com/kr4phy/r2pipe-go"
)

func main() {
    r2p, err := r2pipe.NewApiPipe("/bin/ls")
    if err != nil {
        panic(err)
    }
    defer r2p.Close()

    result, err := r2p.Cmd("pd 10 @ entry0")
    if err != nil {
        panic(err)
    }
    fmt.Println(result)
}
```

Build with: `go build -tags r2api`

## API Reference

### Pipe Creation

| Function | Description |
|----------|-------------|
| `NewPipe(file string)` | Create a pipe by spawning radare2 process |
| `NewNativePipe(file string)` | Create a pipe using dynamic library loading |
| `NewApiPipe(file string)` | Create a pipe using direct C API (requires `r2api` tag) |

### Command Execution

| Method | Description |
|--------|-------------|
| `Cmd(cmd string)` | Execute a command and return output |
| `Cmdf(format string, args ...any)` | Execute a formatted command |
| `Cmdj(cmd string)` | Execute command and parse JSON output |
| `CmdjStruct(cmd string, out any)` | Execute command and unmarshal JSON into struct |
| `Cmdjf(format string, args ...any)` | Formatted command with JSON parsing |
| `CmdjfStruct(format string, out any, args ...any)` | Formatted command with struct unmarshaling |

### I/O Operations

| Method | Description |
|--------|-------------|
| `Read(p []byte)` | Read from stdout |
| `Write(p []byte)` | Write to stdin |
| `ReadErr(p []byte)` | Read from stderr |

### Lifecycle

| Method | Description |
|--------|-------------|
| `Close()` | Gracefully close the pipe |
| `ForceClose()` | Force close the pipe (sends `q!`) |

## Running Tests

```bash
# Run basic tests (requires radare2 in PATH)
go test -v -run TestCmd

# Run native tests (requires libr_core.so installed)
go test -v -run TestNativeCmd

# Run all tests
go test -v

# Run API tests (requires radare2 development libraries)
go test -v -tags r2api -run TestApiCmd
```

## Build Tags

| Tag | Description |
|-----|-------------|
| `r2api` | Enable direct C API linking (requires radare2-dev) |

## Troubleshooting

### "failed to open libr_core.so"

Make sure radare2 libraries are in your library path:

```bash
# Add radare2 library path
echo "/usr/local/lib" | sudo tee /etc/ld.so.conf.d/radare2.conf
sudo ldconfig
```

### "radare2: command not found"

Install radare2 or add it to your PATH:

```bash
# Install from source
git clone https://github.com/radareorg/radare2
cd radare2
./configure --prefix=/usr/local
make -j$(nproc)
sudo make install
```

## Attribution

This project is forked from [radareorg/r2pipe-go](https://github.com/radareorg/r2pipe-go) and includes:

- Bug fixes for memory leaks and variable shadowing
- Compatibility updates for radare2 6.x
- Modern Go idioms (Go 1.23+)
- Improved error handling and documentation

## License

MIT License - see [LICENSE](LICENSE) file for details.
