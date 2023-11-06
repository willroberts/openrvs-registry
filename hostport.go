package registry

import "fmt"

// Hostport represents an IP+Port combo to be used as a unique server ID.
type Hostport string

// NewHostport combines an IP and Port into a unique server ID.
func NewHostport(ip string, port int) Hostport {
	return Hostport(fmt.Sprintf("%s:%d", ip, port))
}
