package example

import (
	"net"
	"time"

	"github.com/kklab-com/gone-core/channel"
	gone_tcp "github.com/kklab-com/gone-tcp"
)

type Server struct {
}

func (k *Server) Start(localAddr net.Addr) {
	bootstrap := channel.NewServerBootstrap()
	bootstrap.ChannelType(&gone_tcp.ServerChannel{})
	bootstrap.ChildHandler(channel.NewInitializer(func(ch channel.Channel) {
		ch.Pipeline().AddLast("DECODE_HANDLER", NewDecodeHandler())
		ch.Pipeline().AddLast("HANDLER", &ServerChildHandler{})
	}))

	ch := bootstrap.Bind(localAddr).Sync().Channel()
	go func() {
		time.Sleep(time.Minute * 1)
		ch.Close()
	}()

	ch.CloseFuture().Sync()
}
