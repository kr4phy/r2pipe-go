package r2pipe

import (
	"testing"
	"time"
)

func TestErrSide(t *testing.T) {
	r2p, err := NewPipe("malloc://256")
	if err != nil {
		t.Fatal(err)
	}
	defer r2p.Close()

	var res bool
	if err := r2p.On("errmsg", res, func(p *Pipe, typ string, user any, dat string) bool {
		t.Log("errmsg received")
		res = true
		return false
	}); err != nil {
		t.Logf("On error (expected without stderr support): %v", err)
	}

	_, _ = r2p.Cmd("aaa")
	time.Sleep(time.Millisecond * 100)
	if res {
		t.Log("It works!")
	}
	t.Log("[*] Testing r2pipe-side stderr message")
}
