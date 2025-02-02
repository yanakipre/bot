package container

import (
	"fmt"
)

type Address struct {
	host  string
	port  string
	proto string
}

func NewAddress(host, port, proto string) Address {
	return Address{
		host:  host,
		port:  port,
		proto: proto,
	}
}

func (a Address) String() string {
	return fmt.Sprintf("%s://%s:%s", a.proto, a.host, a.port)
}

func (a Address) StringWithoutProto() string {
	return fmt.Sprintf("%s:%s", a.host, a.port)
}
