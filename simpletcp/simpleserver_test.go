package simpletcp

import (
	"net"
	"testing"
	"time"

	"github.com/kklab-com/gone-core/channel"
	buf "github.com/kklab-com/goth-bytebuf"
	"github.com/kklab-com/goth-kklogger"
	"github.com/stretchr/testify/assert"
)

type testServerHandler struct {
	channel.DefaultHandler
}

func (h *testServerHandler) Read(ctx channel.HandlerContext, obj interface{}) {
	ctx.Channel().Write(obj)
}

type testClientHandler struct {
	channel.DefaultHandler
	num    int32
	active int
	read   int
}

func (h *testClientHandler) Active(ctx channel.HandlerContext) {
	h.active++
	ctx.Channel().Write(buf.EmptyByteBuf().WriteInt32(h.num))
}

func (h *testClientHandler) Read(ctx channel.HandlerContext, obj interface{}) {
	if obj.(buf.ByteBuf).ReadInt32() == h.num {
		h.num++
		h.read++
		ctx.Channel().Disconnect()
	}
}

func TestServer_Start(t *testing.T) {
	kklogger.SetLogLevel("DEBUG")
	server := NewServer(&testServerHandler{})
	sch := server.Start(&net.TCPAddr{IP: nil, Port: 18082})
	assert.NotNil(t, sch)
	for i := 0; i < 10; i++ {
		go func(t *testing.T) {
			tcHandler := &testClientHandler{}
			client := NewClient(tcHandler)
			client.AutoReconnect = func() bool {
				return tcHandler.active < 10
			}

			cch := client.Start(&net.TCPAddr{IP: nil, Port: 18082})
			assert.NotNil(t, cch)
			time.Sleep(time.Second * 2)
			assert.Equal(t, 10, tcHandler.read)
			assert.Equal(t, 10, tcHandler.active)
		}(t)
	}

	go func(t *testing.T) {
		tcHandler := &testClientHandler{}
		client := NewClient(tcHandler)
		client.AutoReconnect = func() bool {
			return tcHandler.active < 10
		}

		cch := client.Start(&net.TCPAddr{IP: nil, Port: 18082})
		assert.NotNil(t, cch)
		time.Sleep(time.Second * 2)
		assert.Equal(t, 10, tcHandler.read)
		assert.Equal(t, 10, tcHandler.active)
		server.Stop().Sync()
	}(t)

	server.Channel().CloseFuture().Sync()
}
