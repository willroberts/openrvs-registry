package v2

import (
	"fmt"
	"net"
)

// Handler handles an incoming UDP request from the given address, with any
// associated data or error.
type Handler func(addr *net.UDPAddr, data []byte, err error)

// HandleUDP binds the given handler to incoming requests on the given UDP
// port. This function is blocking, but can be run as a goroutine and stopped
// by sending to stopCh.
func HandleUDP(port int, h Handler, stopCh chan struct{}) error {
	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf(":%d", port))
	if err != nil {
		return err
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return err
	}
	defer conn.Close()

	buf := make([]byte, 4096)
	select {
	case _ = <-stopCh:
		break
	default:
		n, addr, err := conn.ReadFromUDP(buf) // Blocking
		go h(addr, buf[0:n], err)
	}

	return nil
}
