package udp

import (
	"fmt"
	"net"
)

type UDPHandler func(addr *net.UDPAddr, data []byte, err error, stopCh chan struct{})

// HandleUDP binds the given handler to incoming requests on the given UDP
// port. This function is blocking, but can be run as a goroutine and stopped
// by sending to stopCh.
func HandleUDP(port int, handler UDPHandler, stopCh chan struct{}) error {
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
		go handler(addr, buf[0:n], err, stopCh)
	}

	return nil
}
