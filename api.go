//go:build cgo && r2api

// radare - LGPL - Copyright 2021 - pancake

package r2pipe

// #cgo CFLAGS: -I/usr/local/include/libr
// #cgo CFLAGS: -I/usr/local/include/libr/sdb
// #cgo LDFLAGS: -L/usr/local/lib -lr_core
// #cgo pkg-config: r_core
// #include <stdio.h>
// #include <stdlib.h>
// extern void r_core_free(void *);
// extern void *r_core_new(void);
// extern char *r_core_cmd_str(void*, const char *);
import "C"

import (
	"unsafe"
)

// ApiCmd executes a radare2 command using the C API and returns the result.
func (r2p *Pipe) ApiCmd(cmd string) (string, error) {
	ccmd := C.CString(cmd)
	defer C.free(unsafe.Pointer(ccmd))
	res := C.r_core_cmd_str(r2p.Core, ccmd)
	goRes := C.GoString(res)
	C.free(unsafe.Pointer(res))
	return goRes, nil
}

// ApiClose closes the radare2 core instance.
func (r2p *Pipe) ApiClose() error {
	C.r_core_free(unsafe.Pointer(r2p.Core))
	r2p.Core = nil
	return nil
}

// NewApiPipe creates a new Pipe using the radare2 C API.
// This requires radare2 development libraries to be installed.
func NewApiPipe(file string) (*Pipe, error) {
	r2 := C.r_core_new()
	r2p := &Pipe{
		File: file,
		Core: r2,
		cmd: func(r2p *Pipe, cmd string) (string, error) {
			return r2p.ApiCmd(cmd)
		},
		close: func(r2p *Pipe) error {
			return r2p.ApiClose()
		},
	}
	if file != "" {
		_, _ = r2p.ApiCmd("o " + file)
	}
	return r2p, nil
}
