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

	reg := NewRegistry(RegistryConfig{})
	stopCh := make(chan struct{})
	go reg.HandleUDP(9999, testHandler, stopCh)
	stopCh <- struct{}{}
}
