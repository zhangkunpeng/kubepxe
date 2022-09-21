package server

import (
	"fmt"
	"github.com/insomniacslk/dhcp/dhcpv4/server4"
	//"github.com/zhangkunpeng/kubepxe/dhcp/option"
	"golang.org/x/net/ipv4"
	"net"
	"sync"
)

type Server4 struct {
	conn *ipv4.PacketConn
	ifi  *net.Interface
	addr net.UDPAddr

	//handlers option.Handlers

	hlock sync.RWMutex
}

func (s *Server4) LocalAddr() net.Addr {
	return s.conn.LocalAddr()
}

func (s *Server4) Start(stop chan struct{}) {
	log.Printf("Listen %s", s.conn.LocalAddr())

	for {
		b := *bufpool.Get().(*[]byte)
		b = b[:MaxDatagram] //Reslice to max capacity in case the buffer in pool was resliced smaller

		n, oob, peer, err := s.conn.ReadFrom(b)
		if err != nil {
			log.Printf("Error reading from connection: %v", err)
			return
		}
		go s.HandleMessage(b[:n], oob, peer.(*net.UDPAddr))
	}
}

func (s *Server4) Stop() error {
	return s.conn.Close()
}

//func (s *Server4) SetHandler(name uint8, h option.Handler) {
//	s.hlock.Lock()
//	defer s.hlock.Unlock()
//	if h == nil {
//		delete(s.handlers, name)
//	} else {
//		s.handlers[name] = h
//	}
//}

func (s *Server4) HandleMessage(buf []byte, oob *ipv4.ControlMessage, _peer net.Addr) {

}

func NewServer4(addr *net.UDPAddr, opt Options) (*Server4, error) {
	s := &Server4{addr: *addr}
	c, err := server4.NewIPv4UDPConn(addr.Zone, addr)
	if err != nil {
		return nil, err
	}
	s.conn = ipv4.NewPacketConn(c)
	if len(addr.Zone) == 0 {
		err = s.conn.SetControlMessage(ipv4.FlagInterface, true)
		if err != nil {
			return nil, err
		}
	} else {
		s.ifi, err = net.InterfaceByName(addr.Zone)
		if err != nil {
			return nil, fmt.Errorf("DHCPv4: Listen could not find interface %s: %v", addr.Zone, err)
		}
	}
	if addr.IP.IsMulticast() {
		err = s.conn.JoinGroup(s.ifi, addr)
		if err != nil {
			return nil, err
		}
	}
	return s, nil
}
