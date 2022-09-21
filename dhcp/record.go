package dhcp

import (
	"net"
	"time"
)

type Record struct {
	SN      string
	Address net.IPNet
	Expires time.Time
	Mac     string
}

var RecordHandle = _recordHandler

var _recordHandler = func(r Record) {

}
