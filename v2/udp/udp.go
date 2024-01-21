package udp

import "net"

type UDPServer interface {
	AddListener(port int, handler func(net.Addr, []byte))
}
