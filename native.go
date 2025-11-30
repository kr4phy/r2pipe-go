//go:build cgo && (linux || darwin) && !r2api

// radare - LGPL - Copyright 2017 - pancake

package r2pipe

// #cgo linux LDFLAGS: -ldl
// #include <stdio.h>
// #include <dlfcn.h>
// #include <stdlib.h>
// #ifndef RTLD_NOW
// #define RTLD_NOW 2
// #endif
// void *gor_core_new(void *f) {
// 	void *(*rcn)();
// 	rcn = (void *(*)())f;
// 	return rcn();
// }
//
// void gor_core_free(void *f, void *arg) {
// 	void (*fr)(void *);
// 	fr = (void (*)(void *))f;
// 	fr(arg);
// }
//
// char *gor_core_cmd_str(void *f, void *arg, char *arg2) {
// 	char *(*cmdstr)(void *, char *);
// 	cmdstr = (char *(*)(void *, char *))f;
// 	return cmdstr(arg, arg2);
// }
import "C"

import (
	"fmt"
	"runtime"
	"unsafe"
)

// Ptr is an alias for unsafe.Pointer for convenience.
type Ptr = unsafe.Pointer

var (
	lib         Ptr
	rCoreNew    func() Ptr
	rCoreFree   func(Ptr)
	rCoreCmdStr func(Ptr, string) string
)

// DL represents a dynamically loaded library.
type DL struct {
	handle unsafe.Pointer
	name   string
}

func dlOpen(path string) (*DL, error) {
	var ret DL
	switch runtime.GOOS {
	case "darwin":
		path = path + ".dylib"
	case "windows":
		path = path + ".dll"
	default:
		path = path + ".so" // linux/bsds
	}
	cpath := C.CString(path)
	if cpath == nil {
		return nil, fmt.Errorf("failed to allocate C string for path")
	}
	defer C.free(unsafe.Pointer(cpath))

	// Use RTLD_NOW for immediate symbol resolution
	ret.handle = C.dlopen(cpath, C.RTLD_NOW)
	ret.name = path
	if ret.handle == nil {
		errStr := C.GoString(C.dlerror())
		return nil, fmt.Errorf("failed to open %s: %s", path, errStr)
	}
	return &ret, nil
}

func dlSym(dl *DL, name string) (unsafe.Pointer, error) {
	cname := C.CString(name)
	if cname == nil {
		return nil, fmt.Errorf("failed to allocate C string for symbol name")
	}
	defer C.free(unsafe.Pointer(cname))

	handle := C.dlsym(dl.handle, cname)
	if handle == nil {
		errStr := C.GoString(C.dlerror())
		return nil, fmt.Errorf("failed to load '%s' from '%s': %s", name, dl.name, errStr)
	}
	return handle, nil
}

// NativeLoad loads the radare2 native library dynamically.
func NativeLoad() error {
	if lib != nil {
		return nil
	}

	dl, err := dlOpen("libr_core")
	if err != nil {
		return err
	}
	lib = dl.handle

	handle1, err := dlSym(dl, "r_core_new")
	if err != nil {
		return err
	}
	rCoreNew = func() Ptr {
		return Ptr(C.gor_core_new(handle1))
	}

	handle2, err := dlSym(dl, "r_core_free")
	if err != nil {
		return err
	}
	rCoreFree = func(p Ptr) {
		C.gor_core_free(handle2, unsafe.Pointer(p))
	}

	handle3, err := dlSym(dl, "r_core_cmd_str")
	if err != nil {
		return err
	}
	rCoreCmdStr = func(p Ptr, s string) string {
		a := C.CString(s)
		defer C.free(unsafe.Pointer(a))
		b := C.gor_core_cmd_str(handle3, unsafe.Pointer(p), a)
		goRes := C.GoString(b)
		C.free(unsafe.Pointer(b))
		return goRes
	}
	return nil
}

// NativeCmd executes a radare2 command using the native library.
func (r2p *Pipe) NativeCmd(cmd string) (string, error) {
	res := rCoreCmdStr(r2p.Core, cmd)
	return res, nil
}

// NativeClose closes the native radare2 core instance.
func (r2p *Pipe) NativeClose() error {
	rCoreFree(r2p.Core)
	r2p.Core = nil
	return nil
}

// NewNativePipe creates a new Pipe using dynamic library loading.
// This loads libr_core at runtime without requiring compile-time linking.
func NewNativePipe(file string) (*Pipe, error) {
	if err := NativeLoad(); err != nil {
		return nil, err
	}
	r2 := rCoreNew()
	r2p := &Pipe{
		File: file,
		Core: r2,
		cmd: func(r2p *Pipe, cmd string) (string, error) {
			return r2p.NativeCmd(cmd)
		},
		close: func(r2p *Pipe) error {
			return r2p.NativeClose()
		},
	}
	if file != "" {
		_, _ = r2p.NativeCmd("o " + file)
	}
	return r2p, nil
}
