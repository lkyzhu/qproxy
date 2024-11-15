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
	n, err := self.stream.Read(b)
	if err != nil {
		return 0, err
	}

	if self.cipher != nil {
		self.cipher.XORKeyStream(b[:n], b[:n])
	}

	return n, nil
}

func (self *StreamConn) Write(b []byte) (int, error) {
	if self.cipher != nil {
		self.cipher.XORKeyStream(b[:], b[:])
		return self.stream.Write(b)
	} else {
		return self.stream.Write(b)
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