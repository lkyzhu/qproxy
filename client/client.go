package client

import (
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"io/ioutil"
	"log"
	"net"
	"time"

	"github.com/lkyzhu/qproxy/client/conf"
	"github.com/lkyzhu/qproxy/utils"
	"github.com/quic-go/quic-go"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/pkcs12"
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
	log.Printf("local server started ^_^\n")

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

	defer func() {
		str.Close()
		str.CancelRead(0)
		str.CancelWrite(0)
	}()

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

	cert, err := self.loadAgentCert()
	if err != nil {
		log.Fatalf("load agent cert fail, err:%v\n", err.Error())
	}

	tlsConf := &tls.Config{
		Certificates:       []tls.Certificate{*cert},
		ClientSessionCache: tls.NewLRUClientSessionCache(100),
		InsecureSkipVerify: true,
	}

	conn, err := quic.DialAddr(ctx, self.Config.Agent.Addr, tlsConf, &quic.Config{KeepAlivePeriod: 10 * time.Second})
	if err != nil {
		log.Fatalf("quic.transport dial remote[%v] fail, err:%v\n", self.Config.Local.Addr, err.Error())
	}
	self.conn = conn
	log.Printf("connect remote server success\n")

	return nil
}

func (self *Client) loadAgentCert() (*tls.Certificate, error) {
	cert := tls.Certificate{}

	if self.Config.Agent.Cert.CertData != "" && self.Config.Agent.Cert.KeyData != "" {
		c, err := tls.X509KeyPair([]byte(self.Config.Agent.Cert.CertData), []byte(self.Config.Agent.Cert.KeyData))
		if err != nil {
			log.Printf("pair tls cert fail, err:%v\n", err.Error())
			return nil, err
		}

		cert = c
	} else if self.Config.Agent.Cert.Pfx != "" {
		data, err := ioutil.ReadFile(self.Config.Agent.Cert.Pfx)
		if err != nil {
			log.Printf("load pfx cert[%v] fail, err:%v\n", self.Config.Agent.Cert.Pfx, err.Error())
			return nil, err
		}
		private, certificate, err := pkcs12.Decode(data, self.Config.Agent.Cert.Pwd)
		if err != nil {
			log.Printf("load decode cert[%v] fail, err:%v\n", self.Config.Agent.Cert.Pfx, err.Error())
			return nil, err
		}

		if pData, ok := private.([]byte); ok {
			c, err := tls.X509KeyPair(certificate.Raw, pData)
			if err != nil {
				log.Printf("pair tls cert[%v] fail, err:%v\n", self.Config.Agent.Cert.Pfx, err.Error())
				return nil, err
			}

			cert = c
		} else {
			log.Printf("pair tls cert[%v] fail, err:%v\n", self.Config.Agent.Cert.Pfx, "private is invalid")
			return nil, err
		}
	} else {
		c, err := tls.LoadX509KeyPair(self.Config.Agent.Cert.Cert, self.Config.Agent.Cert.Key)
		if err != nil {
			log.Printf("load cert[%v/%v] fail, err:%v\n", self.Config.Agent.Cert.Cert, self.Config.Agent.Cert.Key, err.Error())
			return nil, err
		}

		cert = c
	}

	return &cert, nil
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
