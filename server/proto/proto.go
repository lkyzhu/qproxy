package proto

import (
	"fmt"
	"net"

	"github.com/lkyzhu/qproxy/server/proto/socks5"
)

type Server interface {
	ServeConn(conn net.Conn) error
}

func NewServer(pType string) (Server, error) {
	switch pType {
	case "socks5":
		return socks5.NewServer(), nil
	default:
		return nil, fmt.Errorf("not support proto type")
	}
}
