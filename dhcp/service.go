package dhcp

import (
	"github.com/zhangkunpeng/kubepxe/dhcp/server"
	"net"
)

type Service interface {
	Start(stop chan struct{})

	Stop() error

	LocalAddr() net.Addr
}

func NewService(addr string, opt server.Options) (Service, error) {
	udp, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return nil, err
	}
	if udp.IP.To4() != nil {
		return server.NewServer4(udp, opt)
	}
	return nil, nil
}
