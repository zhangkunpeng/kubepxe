package dhcp

import (
	"net"

	"github.com/coredhcp/coredhcp/config"
	"github.com/insomniacslk/dhcp/dhcpv4"
	"github.com/insomniacslk/dhcp/dhcpv6"
)

const (
	DefaultServerPortV4 = dhcpv4.ServerPort
	DefaultServerPortV6 = dhcpv6.DefaultServerPort
)

type Config config.Config

type ServerConfig config.ServerConfig

type PluginConfig config.PluginConfig

func NewConfig(ver int, sc *ServerConfig) *Config {
	conf := &Config{}
	switch ver {
	case 4:
		conf.Server4 = (*config.ServerConfig)(sc)
	case 6:
		conf.Server6 = (*config.ServerConfig)(sc)
	}
	return conf
}

func NewServerConfig(listeners []net.UDPAddr, plugins []PluginConfig) *ServerConfig {
	pls := make([]config.PluginConfig, len(plugins))
	for _, p := range plugins {
		pls = append(pls, config.PluginConfig(p))
	}
	sc := &ServerConfig{
		Addresses: listeners,
		Plugins:   pls,
	}
	return sc
}
