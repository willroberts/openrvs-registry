package udp

import (
	"log"
	"net"
	"testing"
)

func testHandler(addr *net.UDPAddr, data []byte, err error, stopCh chan struct{}) {
	log.Println("Received UDP from", addr.IP.String())
	if err != nil {
		log.Println("UDP error:", err)
		stopCh <- struct{}{}
		return
	}
	// doSomeWork()
	stopCh <- struct{}{}
}

func TestUDP(t *testing.T) {
	stopCh := make(chan struct{})
	go HandleUDP(9999, testHandler, stopCh)
	stopCh <- struct{}{}
}
