package tcp

import (
	"fmt"
	"net"

	"github.com/kklab-com/gone-core/channel"
)

type Channel struct {
	channel.DefaultNetChannel
}

var ErrNotTCPAddr = fmt.Errorf("not tcp addr")

func (c *Channel) UnsafeConnect(localAddr net.Addr, remoteAddr net.Addr) error {
	if remoteAddr == nil {
		return channel.ErrNilObject
	}

	if _, ok := remoteAddr.(*net.TCPAddr); !ok {
		return ErrNotTCPAddr
	}

	if localAddr != nil {
		if _, ok := localAddr.(*net.TCPAddr); !ok {
			return ErrNotTCPAddr
		}
	}

	return c.DefaultNetChannel.UnsafeConnect(localAddr, remoteAddr)
}
