// Example demonstrating basic r2pipe-go usage
//
// This example shows how to:
// - Create a pipe to radare2
// - Write and read data
// - Execute commands
// - Use formatted commands
// - Work with hex dumps
package main

import (
	"fmt"
	"os"

	r2pipe "github.com/kr4phy/r2pipe-go"
)

func main() {
	fmt.Println("=== r2pipe-go Example ===")
	fmt.Println()

	// Example 1: Basic memory operations
	fmt.Println("1. Basic Memory Operations")
	fmt.Println("--------------------------")
	basicMemoryExample()
	fmt.Println()

	// Example 2: Binary analysis
	fmt.Println("2. Binary Analysis")
	fmt.Println("------------------")
	binaryAnalysisExample()
	fmt.Println()

	fmt.Println("=== All examples completed successfully! ===")
}

// basicMemoryExample demonstrates basic memory read/write operations
func basicMemoryExample() {
	// Create a new pipe with malloc (memory allocation)
	r2p, err := r2pipe.NewPipe("malloc://256")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating pipe: %v\n", err)
		os.Exit(1)
	}
	defer r2p.Close()

	// Write some data
	_, err = r2p.Cmd("w Hello r2pipe-go!")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing: %v\n", err)
		os.Exit(1)
	}

	// Read back the data as string
	result, err := r2p.Cmd("ps")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Written string: %s\n", result)

	// Get radare2 version
	version, err := r2p.Cmd("?V")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting version: %v\n", err)
		os.Exit(1)
	}
	// Only print first line of version
	if len(version) > 40 {
		version = version[:40]
	}
	fmt.Printf("Radare2: %s\n", version)

	// Use formatted command for hex dump
	result, err = r2p.Cmdf("px %d", 16)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Hex dump (16 bytes):\n%s\n", result)
}

// binaryAnalysisExample demonstrates binary analysis capabilities
func binaryAnalysisExample() {
	// Analyze /bin/ls
	r2p, err := r2pipe.NewPipe("/bin/ls")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating pipe: %v\n", err)
		os.Exit(1)
	}
	defer r2p.Close()

	// Get binary info
	info, err := r2p.Cmd("i")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting info: %v\n", err)
		os.Exit(1)
	}
	// Print first few lines
	lines := 0
	for i, c := range info {
		if c == '\n' {
			lines++
			if lines >= 5 {
				info = info[:i]
				break
			}
		}
	}
	fmt.Printf("Binary info:\n%s\n...\n", info)

	// Disassemble entry point
	disasm, err := r2p.Cmd("pd 5 @ entry0")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error disassembling: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Entry point disassembly:\n%s\n", disasm)
}
