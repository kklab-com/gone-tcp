package example

import (
	"net"
	"testing"
	"time"

	"github.com/kklab-com/gone-core/channel"
	"github.com/kklab-com/gone-tcp/tcp"
	"github.com/kklab-com/goth-kklogger"
	"github.com/kklab-com/goth-kkutil/buf"
	"github.com/kklab-com/goth-kkutil/concurrent"
)

func TestServer_Start(t *testing.T) {
	kklogger.SetLogLevel("DEBUG")
	bootstrap := channel.NewServerBootstrap()
	bootstrap.ChannelType(&tcp.ServerChannel{})
	bootstrap.Handler(channel.NewInitializer(func(ch channel.Channel) {
		ch.Pipeline().AddLast("INDICATE_HANDLER_INBOUND", &channel.IndicateHandlerInbound{})
		ch.Pipeline().AddLast("INDICATE_HANDLER_OUTBOUND", &channel.IndicateHandlerOutbound{})
	}))

	bootstrap.ChildHandler(channel.NewInitializer(func(ch channel.Channel) {
		ch.Pipeline().AddLast("INDICATE_HANDLER_INBOUND", &channel.IndicateHandlerInbound{})
		ch.Pipeline().AddLast("DECODE_HANDLER", NewDecodeHandler())
		ch.Pipeline().AddLast("HANDLER", &ServerChildHandler{})
		ch.Pipeline().AddLast("INDICATE_HANDLER_OUTBOUND", &channel.IndicateHandlerOutbound{})
	}))

	ch := bootstrap.Bind(&net.TCPAddr{IP: nil, Port: 18082}).Sync().Channel()
	go func() {
		time.Sleep(time.Second * 3)
		ch.Close()
	}()

	go func() {
		bootstrap := channel.NewBootstrap()
		bootstrap.ChannelType(&tcp.Channel{})
		bootstrap.Handler(channel.NewInitializer(func(ch channel.Channel) {
			ch.Pipeline().AddLast("INDICATE_HANDLER_INBOUND", &channel.IndicateHandlerInbound{})
			ch.Pipeline().AddLast("DECODE_HANDLER", NewDecodeHandler())
			ch.Pipeline().AddLast("HANDLER", &ClientHandler{})
			ch.Pipeline().AddLast("INDICATE_HANDLER_OUTBOUND", &channel.IndicateHandlerOutbound{})
		}))

		bwg := concurrent.BurstWaitGroup{}
		for i := 0; i < 10; i++ {
			bwg.Add(1)
			go func(i int) {
				ch := bootstrap.Connect(nil, &net.TCPAddr{IP: nil, Port: 18082}).Sync().Channel()
				ch.Write(buf.NewByteBuf([]byte("o12b32c49")))
				time.Sleep(time.Millisecond * 10)
				ch.Write(buf.NewByteBuf([]byte("a42d22e41")))
				time.Sleep(time.Millisecond * 10)
				if i%2 == 0 {
					ch.Disconnect()
				}

				bwg.Done()
			}(i)
		}

		bwg.Wait()
		time.Sleep(time.Second * 111111)
		nch := bootstrap.Connect(nil, &net.TCPAddr{IP: nil, Port: 18082}).Sync().Channel()
		nch.Write(buf.NewByteBuf([]byte("ccc")))
	}()

	ch.CloseFuture().Sync()
}