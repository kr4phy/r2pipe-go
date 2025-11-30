// radare - LGPL - Copyright 2015 - nibble

package r2pipe

import (
	"testing"
)

type Offset struct {
	Offset  uint
	Current bool
}

func TestCmd(t *testing.T) {
	t.Log("[*] Testing r2 spawn pipe")
	r2p, err := NewPipe("malloc://256")
	if err != nil {
		t.Fatal(err)
	}
	defer r2p.Close()

	check := "Hello World"

	_, err = r2p.Cmd("w " + check)
	if err != nil {
		t.Fatal(err)
	}
	buf, err := r2p.Cmd("ps")
	if err != nil {
		t.Fatal(err)
	}
	if buf != check {
		t.Errorf("buf=%v; want=%v", buf, check)
	}

	offset := Offset{}
	if err := r2p.CmdjStruct("sj ~{0}", &offset); err != nil {
		t.Logf("CmdjStruct error (expected in some r2 versions): %v", err)
	}

	if err := r2p.CmdjfStruct("sj ~{%d}", &offset, 0); err != nil {
		t.Logf("CmdjfStruct error (expected in some r2 versions): %v", err)
	}
}
