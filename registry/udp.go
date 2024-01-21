package registry

import (
	"fmt"
	"net"
)

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
