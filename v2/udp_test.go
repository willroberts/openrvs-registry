package v2

import (
	"net"
	"testing"
)

func TestUDP(t *testing.T) {
	testHandler := func(addr *net.UDPAddr, data []byte, err error) {
		t.Log("Received UDP from", addr.IP.String())
		if err != nil {
			t.Log("UDP error:", err)
		}
	}

	stopCh := make(chan struct{})
	go HandleUDP(9999, testHandler, stopCh)
	stopCh <- struct{}{}
}
