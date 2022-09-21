package server

import "net"

type Options struct {
	// IP version: 4 or 6
	Version int

	// start ip address for allocating
	Start net.IPNet
	//Size: is the total count of addresses for IPV4
	// For Ipv6, the total count is 2^Size
	Size int

	//Listen Address of Service
	Listen net.UDPAddr
}
