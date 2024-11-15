package server

import (
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/lkyzhu/qproxy/server/conf"
	"github.com/lkyzhu/qproxy/server/proto"
	"github.com/lkyzhu/qproxy/utils"
	"github.com/quic-go/quic-go"
	"github.com/sirupsen/logrus"
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
		logrus.Fatalf("resolve udp addr[%s] fail, err:%v\n", self.Config.Addr, err.Error())
	}

	conn, err := net.ListenUDP("udp4", addr)
	if err != nil {
		logrus.Fatalf("listen udp fail, err:%v\n", err.Error())
	}

	transport := quic.Transport{
		Conn: conn,
	}

	cert, err := tls.LoadX509KeyPair(self.Config.Cert.Cert, self.Config.Cert.Key)
	if err != nil {
		logrus.Fatalf("load cert[%v/%v] fail, err:%v\n", self.Config.Cert.Cert, self.Config.Cert.Key, err.Error())
	}

	tlsConf := &tls.Config{
		Certificates: []tls.Certificate{cert},
	}

	quicConf := &quic.Config{
		KeepAlivePeriod: 10 * time.Second,
	}

	listener, err := transport.Listen(tlsConf, quicConf)
	if err != nil {
		logrus.Fatalf("quic.transport listen fail, err:%v\n", err.Error())
	}

	ctx := context.Background()
	for {
		conn, err := listener.Accept(ctx)
		if err != nil {
			logrus.Fatalf("quic.listener accept fail, err:%v\n", err.Error())
		}

		logrus.Printf("accept new conn:%v\n", conn.RemoteAddr().String())
		go self.ServeConn(conn)
	}
}

func (self *Server) ServeConn(conn quic.Connection) error {
	for {
		stream, err := conn.AcceptStream(context.Background())
		if err != nil {
			logrus.Printf("accept new stream fail, err:%v\n", err.Error())

			conn.CloseWithError(5, err.Error())
			return err
		}

		go func() {
			defer stream.Close()

			hello, err := self.ReadHello(stream)
			if err != nil {
				logrus.Printf("accept new stream fail, err:%v\n", err.Error())
				return
			}

			sConn, err := utils.NewStreamConn(conn, stream, []byte(hello.Key), []byte(hello.Nonce))
			if err != nil {
				logrus.WithError(err).Errorf("create new stream conn for stream:%v/%v fail\n", conn.RemoteAddr().String(), stream.StreamID())
				return
			}

			logrus.Printf("accept new stream:%v/%v\n", conn.RemoteAddr().String(), stream.StreamID())
			self.Server.ServeConn(sConn)
		}()
	}

	return nil
}

func (self *Server) ReadHello(stream quic.Stream) (*utils.Hello, error) {
	buf := make([]byte, 1024)
	n, err := stream.Read(buf)
	if err != nil {
		logrus.Errorf("read hello msg fail, err:%v\n", err.Error())
		return nil, err
	}

	buf = buf[:n]

	magic := buf[:utils.MagicLen]
	if !bytes.Equal(utils.Magic, magic) {
		logrus.Errorf("magic[%v] invalid\n", magic)
		return nil, errors.New("invalid magic")
	}

	hello := &utils.Hello{}
	err = utils.Unmarshal(buf[utils.MagicLen:], hello)
	if err != nil {
		logrus.Errorf("unmarshal hello msg fail, err:%v\n", err.Error())
		return nil, err
	}

	resp := &utils.HelloResp{
		Status: utils.StatusSuccess,
	}
	data, err := utils.Marshal(resp)
	if err != nil {
		logrus.Errorf("marshal hello resp fail, err:%v\n", err.Error())
		return nil, err
	}

	buff := make([]byte, utils.MagicLen+len(data))
	buff[0] = utils.Magic[0]
	buff[1] = utils.Magic[1]
	buff[2] = utils.Magic[2]
	buff[3] = utils.Magic[3]
	copy(buff[utils.MagicLen:], data)

	_, err = stream.Write(buff)
	if err != nil {
		logrus.WithError(err).Errorf("write hello to server fail\n")
		return nil, err
	}

	return hello, nil
}

// A wrapper for io.Writer that also logs the message.
type loggingWriter struct{ io.Writer }

func (w loggingWriter) Write(b []byte) (int, error) {
	fmt.Printf("Server: Got '%s'\n", string(b))
	return w.Writer.Write(b)
}
