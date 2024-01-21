package v2

import "fmt"

type Hostport string

func NewHostport(ip string, port int) Hostport {
	return Hostport(fmt.Sprintf("%s:%d", ip, port))
}
