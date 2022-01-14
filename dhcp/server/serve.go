// Copyright 2018-present the CoreDHCP Authors. All rights reserved
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package server

import (
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"golang.org/x/net/ipv4"
	"golang.org/x/net/ipv6"

	"github.com/coredhcp/coredhcp/config"
	"github.com/coredhcp/coredhcp/handler"
	"github.com/coredhcp/coredhcp/logger"
	"github.com/coredhcp/coredhcp/plugins"
	"github.com/insomniacslk/dhcp/dhcpv4/server4"
	"github.com/insomniacslk/dhcp/dhcpv6/server6"
)

var log = logger.GetLogger("server")

type listener6 struct {
	*ipv6.PacketConn
	net.Interface
	handlers []handler.Handler6
	hlock    sync.RWMutex
}

type listener4 struct {
	*ipv4.PacketConn
	net.Interface
	handlers []handler.Handler4
	hlock    sync.RWMutex
}

type listener interface {
	io.Closer
}

// Servers contains state for a running server (with possibly multiple interfaces/listeners)
type Servers struct {
	listeners []listener
	errors    chan error
}

func listen4(a *net.UDPAddr) (*listener4, error) {
	var err error
	l4 := listener4{}
	udpConn, err := server4.NewIPv4UDPConn(a.Zone, a)
	if err != nil {
		return nil, err
	}
	l4.PacketConn = ipv4.NewPacketConn(udpConn)
	var ifi *net.Interface
	if a.Zone != "" {
		ifi, err = net.InterfaceByName(a.Zone)
		if err != nil {
			return nil, fmt.Errorf("DHCPv4: Listen could not find interface %s: %v", a.Zone, err)
		}
		l4.Interface = *ifi
	} else {

		// When not bound to an interface, we need the information in each
		// packet to know which interface it came on
		err = l4.SetControlMessage(ipv4.FlagInterface, true)
		if err != nil {
			return nil, err
		}
	}

	if a.IP.IsMulticast() {
		err = l4.JoinGroup(ifi, a)
		if err != nil {
			return nil, err
		}
	}
	return &l4, nil
}

func listen6(a *net.UDPAddr) (*listener6, error) {
	l6 := listener6{}
	udpconn, err := server6.NewIPv6UDPConn(a.Zone, a)
	if err != nil {
		return nil, err
	}
	l6.PacketConn = ipv6.NewPacketConn(udpconn)
	var ifi *net.Interface
	if a.Zone != "" {
		ifi, err = net.InterfaceByName(a.Zone)
		if err != nil {
			return nil, fmt.Errorf("DHCPv4: Listen could not find interface %s: %v", a.Zone, err)
		}
		l6.Interface = *ifi
	} else {
		// When not bound to an interface, we need the information in each
		// packet to know which interface it came on
		err = l6.SetControlMessage(ipv6.FlagInterface, true)
		if err != nil {
			return nil, err
		}
	}

	if a.IP.IsMulticast() {
		err = l6.JoinGroup(ifi, a)
		if err != nil {
			return nil, err
		}
	}
	return &l6, nil
}

// Start will start the server asynchronously. See `Wait` to wait until
// the execution ends.
func Start(config *config.Config) (*Servers, error) {
	handlers4, handlers6, err := plugins.LoadPlugins(config)
	if err != nil {
		return nil, err
	}
	srv := Servers{
		errors: make(chan error),
	}

	// listen
	if config.Server6 != nil {
		log.Println("Starting DHCPv6 server")
		for _, addr := range config.Server6.Addresses {
			var l6 *listener6
			l6, err = listen6(&addr)
			if err != nil {
				goto cleanup
			}
			l6.handlers = handlers6
			srv.listeners = append(srv.listeners, l6)
			go func() {
				srv.errors <- l6.Serve()
			}()
		}
	}

	if config.Server4 != nil {
		log.Println("Starting DHCPv4 server")
		for _, addr := range config.Server4.Addresses {
			var l4 *listener4
			l4, err = listen4(&addr)
			if err != nil {
				goto cleanup
			}
			l4.handlers = handlers4
			srv.listeners = append(srv.listeners, l4)
			go func() {
				srv.errors <- l4.Serve()
			}()
		}
	}

	return &srv, nil

cleanup:
	srv.Close()
	return nil, err
}

// Wait waits until the end of the execution of the server.
func (s *Servers) Wait() error {
	log.Debug("Waiting")
	err := <-s.errors
	s.Close()
	return err
}

// Close closes all listening connections
func (s *Servers) Close() {
	for _, srv := range s.listeners {
		if srv != nil {
			srv.Close()
		}
	}
	go func() {
		time.Sleep(2 * time.Second)
		var timer = time.NewTimer(3 * time.Second)
		for {
			select {
			case <-s.errors:
				timer.Reset(3 * time.Second)
			case <-timer.C:
				close(s.errors)
				return
			}
		}
	}()
}

func (s *Servers) Stop() {
	s.errors <- nil
}

func (s *Servers) Update(config *config.Config) error {
	handlers4, handlers6, err := plugins.LoadPlugins(config)
	if err != nil {
		return err
	}
	for _, l := range s.listeners {
		if l4, ok := l.(*listener4); ok {
			func() {
				l4.hlock.Lock()
				defer l4.hlock.Unlock()
				l4.handlers = handlers4
			}()
		}
		if l6, ok := l.(*listener6); ok {
			func() {
				l6.hlock.Lock()
				defer l6.hlock.Unlock()
				l6.handlers = handlers6
			}()
		}
	}
	return nil
}

func (l *listener4) Handlers() []handler.Handler4 {
	var h = make([]handler.Handler4, len(l.handlers))
	l.hlock.RLock()
	defer l.hlock.RUnlock()
	copy(h, l.handlers)
	return h
}

func (l *listener6) Handlers() []handler.Handler6 {
	var h = make([]handler.Handler6, len(l.handlers))
	l.hlock.RLock()
	defer l.hlock.RUnlock()
	copy(h, l.handlers)
	return h
}
