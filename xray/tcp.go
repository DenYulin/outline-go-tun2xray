package xray

import (
	"context"
	"fmt"
	"github.com/xtls/xray-core/common/bytespool"
	"github.com/xtls/xray-core/common/session"
	"github.com/xtls/xray-core/core"
	"io"
	"net"

	t2core "github.com/eycorsican/go-tun2socks/core"
	x2net "github.com/xtls/xray-core/common/net"

	"github.com/xxf098/go-tun2socks-build/pool"
)

type tcpHandler struct {
	ctx context.Context
	v   *core.Instance
}

func NewTCPHandler(ctx context.Context, instance *core.Instance) t2core.TCPConnHandler {
	return &tcpHandler{
		ctx: ctx,
		v:   instance,
	}
}

func (handler *tcpHandler) Handle(conn net.Conn, target *net.TCPAddr) error {
	dest := x2net.DestinationFromAddr(target)
	sid := session.NewID()
	ctx := session.ContextWithID(handler.ctx, sid)
	c, err := core.Dial(ctx, handler.v, dest)
	if err != nil {
		return fmt.Errorf("dial V proxy connection failed: %v", err)
	}
	go handler.relay(conn, c)
	return nil
}

func (handler *tcpHandler) relay(lhs net.Conn, rhs net.Conn) {
	go func() {
		buf := bytespool.Alloc(pool.BufSize)
		io.CopyBuffer(rhs, lhs, buf)
		bytespool.Free(buf)
		lhs.Close()
		rhs.Close()
	}()
	buf := bytespool.Alloc(pool.BufSize)
	io.CopyBuffer(lhs, rhs, buf)
	bytespool.Free(buf)
	lhs.Close()
	rhs.Close()
}
