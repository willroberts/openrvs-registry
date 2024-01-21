package registry

import (
	"fmt"
	"net"
)

// UDPHandler is the signature of a function for processing incoming UDP
// requests. After checking for potential errors, the handler has access to
// the origin/source address, as well as the request bytes.
type UDPHandler func(addr *net.UDPAddr, data []byte, err error)

func (r *registry) HandleUDP(port int, h UDPHandler, stopCh chan struct{}) error {
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
