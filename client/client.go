package client

import (
	"bytes"
	"context"
	"crypto/tls"
	"errors"
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
	defer conn.Close()

	str, err := self.conn.OpenStream()
	if err != nil {
		logrus.WithError(err).Errorf("open stream for conn:%v fail\n", conn.RemoteAddr().String())
		self.conn.CloseWithError(5, err.Error())

		self.RemoteAgentInit()
		return err
	}

	hello, err := self.SendHello(str)
	if err != nil {
		logrus.WithError(err).Errorf("send hello to server for conn:%v fail\n", conn.RemoteAddr().String())
		return err
	}

	sConn, err := utils.NewStreamConn(self.conn, str, []byte(hello.Key), []byte(hello.Nonce))
	if err != nil {
		logrus.WithError(err).Errorf("create stream conn for conn:%v fail\n", conn.RemoteAddr().String())
		return err
	}

	logrus.Debugf("bind stream for conn:%v success, stream:%v\n", conn.RemoteAddr().String(), str.StreamID())
	err = utils.BindIO(conn, sConn)
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

	conn, err := quic.DialAddr(ctx, self.Config.Agent.Addr, tlsConf, &quic.Config{KeepAlivePeriod: 10 * time.Second})
	if err != nil {
		log.Fatalf("quic.transport dial remote[%v] fail, err:%v\n", self.Config.Local.Addr, err.Error())
	}
	self.conn = conn

	return nil
}

func (self *Client) SendHello(str quic.Stream) (*utils.Hello, error) {
	hello := &utils.Hello{
		Version: utils.V1,
		Key:     self.Config.Agent.Key,
		Nonce:   self.Config.Agent.Nonce,
	}

	data, err := utils.Marshal(hello)
	if err != nil {
		logrus.WithError(err).Errorf("marshal Hello msg fail")
		return nil, err
	}

	certs := self.conn.ConnectionState().TLS.PeerCertificates
	if len(certs) == 0 {
		logrus.Errorf("invalid conn, peer.certs is nil\n")
		return nil, errors.New("invalid conn")
	}

	buff := make([]byte, utils.MagicLen+len(data))
	buff[0] = utils.Magic[0]
	buff[1] = utils.Magic[1]
	buff[2] = utils.Magic[2]
	buff[3] = utils.Magic[3]
	copy(buff[utils.MagicLen:], data)

	_, err = str.Write(buff)
	if err != nil {
		logrus.WithError(err).Errorf("write hello to server fail\n")
		return nil, err
	}

	buf := make([]byte, 1024)
	n, err := str.Read(buf)
	if err != nil {
		logrus.WithError(err).Errorf("read hello msg fail, err:%v\n", err.Error())
		return nil, err
	}
	buf = buf[:n]

	magic := buf[:utils.MagicLen]
	if !bytes.Equal(utils.Magic, magic) {
		logrus.Errorf("magic[%v] invalid\n", magic)
		return nil, errors.New("invalid magic")
	}
	resp := &utils.HelloResp{
		Status: utils.StatusSuccess,
	}
	err = utils.Unmarshal(buf[utils.MagicLen:], resp)
	if err != nil {
		logrus.WithError(err).Errorf("marshal hello resp fail, err:%v\n", err.Error())
		return nil, err
	}
	if resp.Status != utils.StatusSuccess {
		logrus.Errorf("status[%v] is not success\n", resp.Status)
		return nil, errors.New("hello fhandshake ail")
	}

	return hello, nil
}

func (self *Client) Close() {
}
