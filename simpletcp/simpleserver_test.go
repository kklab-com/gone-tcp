package simpletcp

import (
	"net"
	"testing"

	"github.com/kklab-com/gone-core/channel"
	buf "github.com/kklab-com/goth-bytebuf"
	concurrent "github.com/kklab-com/goth-concurrent"
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
	wg     concurrent.WaitGroup
}

func (h *testClientHandler) Active(ctx channel.HandlerContext) {
	h.active++
	ctx.Channel().Write(buf.EmptyByteBuf().WriteInt32(h.num))
}

func (h *testClientHandler) Read(ctx channel.HandlerContext, obj interface{}) {
	if obj.(buf.ByteBuf).ReadInt32() == h.num {
		h.wg.Done()
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
	count := 10
	for i := 0; i < count; i++ {
		go func(t *testing.T) {
			tcHandler := &testClientHandler{}
			tcHandler.wg.Add(count)
			client := NewClient(tcHandler)
			client.AutoReconnect = func() bool {
				return tcHandler.active < count
			}

			cch := client.Start(&net.TCPAddr{IP: nil, Port: 18082})
			assert.NotNil(t, cch)
			tcHandler.wg.Wait()
			assert.Equal(t, count, tcHandler.read)
			assert.Equal(t, count, tcHandler.active)
		}(t)
	}

	go func(t *testing.T) {
		tcHandler := &testClientHandler{}
		tcHandler.wg.Add(count)
		client := NewClient(tcHandler)
		client.AutoReconnect = func() bool {
			return tcHandler.active < count
		}

		cch := client.Start(&net.TCPAddr{IP: nil, Port: 18082})
		assert.NotNil(t, cch)
		tcHandler.wg.Wait()
		assert.Equal(t, count, tcHandler.read)
		assert.Equal(t, count, tcHandler.active)
		server.Stop()
	}(t)

	server.Channel().CloseFuture().Sync()
}
