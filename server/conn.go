package server

import (
	"net"
	"time"

	"github.com/quic-go/quic-go"
)

type StreamConn struct {
	stream quic.Stream
	conn   quic.Connection
}

func (self *StreamConn) Read(b []byte) (n int, err error) {
	return self.stream.Read(b)
}

func (self *StreamConn) Write(b []byte) (n int, err error) {
	return self.stream.Write(b)
}

func (self *StreamConn) Close() error {
	return self.stream.Close()
}

func (self *StreamConn) LocalAddr() net.Addr {
	return self.conn.LocalAddr()
}

func (self *StreamConn) RemoteAddr() net.Addr {
	return self.conn.RemoteAddr()
}

func (self *StreamConn) SetDeadline(t time.Time) error {
	return self.stream.SetDeadline(t)
}

func (self *StreamConn) SetReadDeadline(t time.Time) error {
	return self.stream.SetReadDeadline(t)
}

func (self *StreamConn) SetWriteDeadline(t time.Time) error {
	return self.stream.SetWriteDeadline(t)
}
