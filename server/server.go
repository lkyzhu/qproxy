package server

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net"

	"github.com/lkyzhu/qproxy/server/conf"
	"github.com/lkyzhu/qproxy/server/proto"
	"github.com/quic-go/quic-go"
)

type Server struct {
	Config *conf.Config
	Server proto.Server
}

func NewServer(config *conf.Config) *Server {
	return &Server{
		Config: config,
	}
}

func (self *Server) Init() error {
	server, err := proto.NewServer("socks5")
	if err != nil {
		return err
	}
	self.Server = server

	return nil
}

func (self *Server) Run() error {
	if err := self.Init(); err != nil {
		return err
	}

	addr, err := net.ResolveUDPAddr("udp4", self.Config.Addr)
	if err != nil {
		log.Fatalf("resolve udp addr[%s] fail, err:%v\n", self.Config.Addr, err.Error())
	}

	conn, err := net.ListenUDP("udp4", addr)
	if err != nil {
		log.Fatalf("listen udp fail, err:%v\n", err.Error())
	}

	transport := quic.Transport{
		Conn: conn,
	}

	cert, err := tls.LoadX509KeyPair(self.Config.Cert.Cert, self.Config.Cert.Key)
	if err != nil {
		log.Fatalf("load cert[%v/%v] fail, err:%v\n", self.Config.Cert.Cert, self.Config.Cert.Key, err.Error())
	}

	tlsConf := &tls.Config{
		Certificates: []tls.Certificate{cert},
	}

	quicConf := &quic.Config{}
	listener, err := transport.Listen(tlsConf, quicConf)
	if err != nil {
		log.Fatalf("quic.transport listen fail, err:%v\n", err.Error())
	}

	ctx := context.Background()
	for {
		conn, err := listener.Accept(ctx)
		if err != nil {
			log.Fatalf("quic.listener accept fail, err:%v\n", err.Error())
		}

		log.Printf("accept new conn:%v\n", conn.RemoteAddr().String())
		go self.ServeConn(conn)
	}
}

func (self *Server) ServeConn(conn quic.Connection) error {
	for {
		stream, err := conn.AcceptStream(context.Background())
		if err != nil {
			log.Printf("accept new stream fail, err:%v\n", err.Error())
		}

		sConn := &StreamConn{
			stream: stream,
			conn:   conn,
		}

		log.Printf("accept new stream:%v/%v\n", conn.RemoteAddr().String(), stream.StreamID())
		go self.Server.ServeConn(sConn)

	}

	return nil
}

// A wrapper for io.Writer that also logs the message.
type loggingWriter struct{ io.Writer }

func (w loggingWriter) Write(b []byte) (int, error) {
	fmt.Printf("Server: Got '%s'\n", string(b))
	return w.Writer.Write(b)
}
