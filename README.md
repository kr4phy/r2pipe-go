# r2pipe-go

[![Go CI](https://github.com/kr4phy/r2pipe-go/actions/workflows/ci.yml/badge.svg)](https://github.com/kr4phy/r2pipe-go/actions/workflows/ci.yml)
[![GoDoc](https://pkg.go.dev/badge/github.com/kr4phy/r2pipe-go)](https://pkg.go.dev/github.com/kr4phy/r2pipe-go)

Go bindings for [radare2](https://github.com/radareorg/radare2), providing a convenient interface to execute r2 commands from Go programs.

## Compatibility

- Go 1.21+
- radare2 6.x

## Installation

```bash
go get github.com/kr4phy/r2pipe-go
```

## Requirements

- radare2 must be installed and available in your PATH
- For API/Native modes: radare2 development libraries (`libr_core`) must be installed

## Usage

### Basic Example (Spawn Pipe)

```go
package main

import (
    "fmt"

    "github.com/kr4phy/r2pipe-go"
)

func main() {
    r2p, err := r2pipe.NewPipe("malloc://256")
    if err != nil {
        panic(err)
    }
    defer r2p.Close()

    _, err = r2p.Cmd("w Hello World")
    if err != nil {
        panic(err)
    }
    buf, err := r2p.Cmd("ps")
    if err != nil {
        panic(err)
    }
    fmt.Println(buf)
}
```

### Native API Mode

For better performance, use the native API which links directly against radare2 libraries:

```go
package main

import (
    "fmt"

    "github.com/kr4phy/r2pipe-go"
)

func main() {
    r2p, err := r2pipe.NewNativePipe("/bin/ls")
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

## Running Tests

```bash
# Basic tests (requires radare2 in PATH)
go test -v -run TestCmd

# All tests (requires radare2 development libraries)
go test -v
```

## Attribution

This project is forked from [radareorg/r2pipe-go](https://github.com/radareorg/r2pipe-go) and includes bug fixes and compatibility updates for radare2 6.x.

## License

MIT License - see [LICENSE](LICENSE) file for details.
