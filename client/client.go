package client

import (
	"context"
	"crypto/tls"
	"log"
	"net"
	"time"

	"github.com/lkyzhu/qproxy/client/conf"
	"github.com/lkyzhu/qproxy/utils"
	"github.com/quic-go/quic-go"
	"github.com/sirupsen/logrus"
)

type Client struct {
	Config *conf.Config
	conn   quic.Connection
}

func NewClient(config *conf.Config) *Client {
	return &Client{Config: config}
}

func (self *Client) Init() error {
	return nil
}

func (self *Client) Run() error {
	defer self.Close()

	self.RemoteAgentInit()

	self.LocalServerRun()

	return nil
}

func (self *Client) LocalServerRun() error {
	listener, err := net.Listen("tcp", self.Config.Local.Addr)
	if err != nil {
		log.Fatalf("listen addr[%v] fail, err:%v\n", self.Config.Local.Addr, err.Error())
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatalf("accept for addr[%v] fail, err:%v\n", self.Config.Local.Addr, err.Error())
		}

		log.Printf("accept new conn:%v\n", conn.RemoteAddr().String())

		go self.BindStream(conn)

	}

	return nil
}

func (self *Client) BindStream(conn net.Conn) error {
	str, err := self.conn.OpenStream()
	if err != nil {
		logrus.WithError(err).Errorf("open stream for conn:%v fail\n", conn.RemoteAddr().String())
		return err
	}

	logrus.Debugf("bind stream for conn:%v success, stream:%v\n", conn.RemoteAddr().String(), str.StreamID())
	err = utils.BindIO(conn, str)
	if err != nil {
		logrus.WithError(err).Errorf("bind io for conn:%v fail\n", conn.RemoteAddr().String())
		return err
	}

	return nil
}

func (self *Client) RemoteAgentInit() error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	cert, err := tls.LoadX509KeyPair(self.Config.Agent.Cert.Cert, self.Config.Agent.Cert.Key)
	if err != nil {
		log.Fatalf("load cert[%v/%v] fail, err:%v\n", self.Config.Agent.Cert.Cert, self.Config.Agent.Cert.Key, err.Error())
	}

	tlsConf := &tls.Config{
		Certificates:       []tls.Certificate{cert},
		ClientSessionCache: tls.NewLRUClientSessionCache(100),
		InsecureSkipVerify: true,
	}

	conn, err := quic.DialAddr(ctx, self.Config.Agent.Addr, tlsConf, &quic.Config{})
	if err != nil {
		log.Fatalf("quic.transport dial remote[%v] fail, err:%v\n", self.Config.Local.Addr, err.Error())
	}
	self.conn = conn

	return nil
}

func (self *Client) Close() {
}
