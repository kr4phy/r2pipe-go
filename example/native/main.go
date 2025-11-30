//go:build cgo && (linux || darwin) && !r2api

// Example demonstrating native r2pipe-go usage
//
// This example shows how to use the native mode which loads
// radare2 libraries dynamically for better performance.
//
// Requirements:
// - radare2 shared libraries (libr_core.so) must be installed
// - Library path must be configured (e.g., /usr/local/lib in LD_LIBRARY_PATH)
package main

import (
	"fmt"
	"os"

	r2pipe "github.com/kr4phy/r2pipe-go"
)

func main() {
	fmt.Println("=== r2pipe-go Native Mode Example ===")
	fmt.Println()

	// Create a native pipe - this loads libr_core.so dynamically
	r2p, err := r2pipe.NewNativePipe("/bin/ls")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating native pipe: %v\n", err)
		fmt.Fprintf(os.Stderr, "\nMake sure radare2 libraries are installed and in your library path.\n")
		fmt.Fprintf(os.Stderr, "Try: echo '/usr/local/lib' | sudo tee /etc/ld.so.conf.d/radare2.conf && sudo ldconfig\n")
		os.Exit(1)
	}
	defer r2p.Close()

	// Get binary information
	fmt.Println("Binary Information:")
	fmt.Println("-------------------")
	info, err := r2p.Cmd("i")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting info: %v\n", err)
		os.Exit(1)
	}
	// Print first 10 lines
	lines := 0
	for i, c := range info {
		if c == '\n' {
			lines++
			if lines >= 10 {
				info = info[:i]
				break
			}
		}
	}
	fmt.Println(info)
	fmt.Println("...")
	fmt.Println()

	// Disassemble entry point
	fmt.Println("Entry Point Disassembly:")
	fmt.Println("------------------------")
	disasm, err := r2p.Cmd("pd 10 @ entry0")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error disassembling: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(disasm)
	fmt.Println()

	// List sections
	fmt.Println("Sections:")
	fmt.Println("---------")
	sections, err := r2p.Cmd("iS")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error listing sections: %v\n", err)
		os.Exit(1)
	}
	// Print first 15 lines
	lines = 0
	for i, c := range sections {
		if c == '\n' {
			lines++
			if lines >= 15 {
				sections = sections[:i]
				break
			}
		}
	}
	fmt.Println(sections)
	fmt.Println("...")
	fmt.Println()

	fmt.Println("=== Native mode example completed successfully! ===")
}
