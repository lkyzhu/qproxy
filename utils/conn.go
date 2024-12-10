package utils

import (
	"net"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/quic-go/quic-go"
	"golang.org/x/crypto/chacha20"
)

func NewStreamConn(conn quic.Connection, stream quic.Stream, key, nonce []byte) (*StreamConn, error) {
	strConn := &StreamConn{
		stream: stream,
		conn:   conn,
		key:    key,
		nonce:  nonce,
		cipher: nil,
	}

	if len(key) == chacha20.KeySize && len(nonce) == chacha20.NonceSize {
		cipher, err := chacha20.NewUnauthenticatedCipher(key, nonce)
		if err != nil {
			return nil, err
		}
		strConn.cipher = cipher
		logrus.Infof("conn:%v . stream:%v use encrypt mode\n", conn.RemoteAddr().String(), stream.StreamID())
	}

	return strConn, nil
}

type StreamConn struct {
	stream quic.Stream
	conn   quic.Connection
	key    []byte
	nonce  []byte
	cipher *chacha20.Cipher
}

func (self *StreamConn) Read(b []byte) (int, error) {
	if self.cipher != nil {
		buf := make([]byte, len(b))
		n, err := self.stream.Read(buf)
		if err != nil {
			return 0, err
		}

		self.cipher.XORKeyStream(b[:n], buf[:n])
		return n, nil
	} else {
		return self.stream.Read(b)
	}
}

func (self *StreamConn) Write(b []byte) (int, error) {
	if self.cipher != nil {
		n := len(b)
		buf := make([]byte, n)

		self.cipher.XORKeyStream(buf, b[:n])
		return self.stream.Write(buf)
	} else {
		return self.stream.Write(b[:])
	}
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
