package discovery

import "testing"

func TestInstanceAddr(t *testing.T) {
	ins := Instance{Host: "127.0.0.1", Port: 9001}
	if got := ins.Addr(); got != "127.0.0.1:9001" {
		t.Fatalf("unexpected addr: %s", got)
	}
}
