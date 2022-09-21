package dhcp

import (
	"fmt"
	"github.com/coredhcp/coredhcp/plugins/allocators/bitmap"
	"net"
	"testing"
	"time"
)

func TestStart(t *testing.T) {
	//_, prefix, _ := net.ParseCIDR("2222::1/64")
	allocSize := 96
	alloc, err := bitmap.NewBitmapAllocator(net.IPNet{IP: net.ParseIP("2222::1"), Mask: net.CIDRMask(64, 128)}, allocSize)
	fmt.Println(err)
	ip, err := alloc.Allocate(net.IPNet{IP: net.ParseIP("2222::ffff:0:0:1"), Mask: net.CIDRMask(allocSize, 128)})
	fmt.Println(err)
	fmt.Println(ip.String())
	ip, err = alloc.Allocate(net.IPNet{})
	fmt.Println(err)
	fmt.Println(ip.String())
	ip, err = alloc.Allocate(net.IPNet{})
	fmt.Println(err)
	fmt.Println(ip.String())
	ip, err = alloc.Allocate(net.IPNet{})
	fmt.Println(err)
	fmt.Println(ip.String())
	time.Sleep(60 * time.Second)
}
