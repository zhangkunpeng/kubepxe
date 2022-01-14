package dhcp

import (
	"fmt"
	"reflect"
	"sync"

	"github.com/coredhcp/coredhcp/config"
	"github.com/zhangkunpeng/kubepxe/dhcp/server"
	"github.com/zhangkunpeng/kubepxe/logger"
)

var log = logger.GetLogger("dhcp")

type Server struct {
	*server.Servers
	config *Config
}

var servers map[string]*Server
var lock sync.RWMutex

func init() {
	servers = make(map[string]*Server)
}

func (c *Config) addressesEqual(nc *Config) bool {

	var scEqual = func(sc *config.ServerConfig, nsc *config.ServerConfig) bool {
		if sc == nsc {
			return true
		}
		if sc != nil && nsc != nil {
			return reflect.DeepEqual(sc.Addresses, nsc.Addresses)
		}
		return false
	}

	return scEqual(c.Server4, nc.Server4) && scEqual(c.Server6, nc.Server6)
}

func (s *Server) updateHandlers(conf *Config) error {
	err := s.Servers.Update((*config.Config)(conf))
	if err != nil {
		log.Errorf("Update Handlers Failed: %v", err)
		return fmt.Errorf("Update Server Failed")
	}
	return nil
}

func Start(name string, conf *Config) (*Server, error) {
	if srv, ok := servers[name]; ok {
		if srv.config.addressesEqual(conf) {
			log.Infof("server: %s, address not update", name)
			return nil, srv.updateHandlers(conf)
		}
		Close(name)
	}
	log.Infof("Start new dhcp server, named %s", name)
	srvs, err := server.Start((*config.Config)(conf))
	if err != nil {
		log.Error(err)
		return nil, err
	}
	srv := &Server{
		Servers: srvs,
		config:  conf,
	}
	go func() {
		lock.Lock()
		defer lock.Unlock()
		servers[name] = srv
	}()
	return srv, nil
}

func Close(name string) {
	lock.Lock()
	defer lock.Unlock()
	if srv, ok := servers[name]; ok {
		srv.Stop()
		delete(servers, name)
	}
}

func GetServer(name string) *Server {
	lock.RLock()
	defer lock.RUnlock()
	if srv, ok := servers[name]; ok {
		return srv
	}
	return nil
}
