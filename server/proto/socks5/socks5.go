package socks5

import (
	s5 "github.com/lkyzhu/socks5"
	"github.com/lkyzhu/socks5/auth"
	"github.com/lkyzhu/socks5/command"
	"github.com/lkyzhu/socks5/resolve"
)

func NewServer() *s5.Server {
	authMgr := &auth.AuthenticatorMgr{}
	noAuth := auth.NewNoAuthAuthenticator()
	authMgr.Regist(noAuth)
	handler := command.NewHandler(resolve.NewResolver())

	server := s5.NewServer(authMgr, handler)
	return server
}
