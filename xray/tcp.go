package xray

import (
	"context"
	"fmt"
	"github.com/DenYulin/outline-go-tun2xray/pool"
	"github.com/eycorsican/go-tun2socks/common/log"
	"github.com/xtls/xray-core/common/bytespool"
	"github.com/xtls/xray-core/common/session"
	"github.com/xtls/xray-core/core"
	"io"
	"net"

	t2core "github.com/eycorsican/go-tun2socks/core"
	x2net "github.com/xtls/xray-core/common/net"
)

type tcpHandler struct {
	ctx      context.Context
	instance *core.Instance
}

func NewTCPHandler(ctx context.Context, instance *core.Instance) t2core.TCPConnHandler {
	return &tcpHandler{
		ctx:      ctx,
		instance: instance,
	}
}

func (handler *tcpHandler) Handle(conn net.Conn, target *net.TCPAddr) error {
	dest := x2net.DestinationFromAddr(target)
	sid := session.NewID()
	ctx := session.ContextWithID(handler.ctx, sid)
	destConn, err := core.Dial(ctx, handler.instance, dest)
	if err != nil {
		log.Errorf("Dial xray proxy connection failed, error: %s", err.Error())
		return fmt.Errorf("dial xray proxy tcp connection failed: %v", err)
	}
	go handler.relay(conn, destConn)
	return nil
}

func (handler *tcpHandler) relay(leftConn net.Conn, rightConn net.Conn) {
	go func() {
		buf := bytespool.Alloc(pool.BufSize)
		if _, err := io.CopyBuffer(rightConn, leftConn, buf); err != nil {
			log.Errorf("Failed from rightConn copy buffer to leftConn, err: %s", err.Error())
		}
		bytespool.Free(buf)
		if leftConn != nil {
			if err := leftConn.Close(); err != nil {
				log.Errorf("Failed to close left conn, error: %s", err.Error())
			}
		}
		if rightConn != nil {
			if err := rightConn.Close(); err != nil {
				log.Errorf("Failed to close right conn, error: %s", err.Error())
			}
		}
	}()

	buf := bytespool.Alloc(pool.BufSize)
	if _, err := io.CopyBuffer(leftConn, rightConn, buf); err != nil {
		log.Errorf("Failed from leftConn copy buffer to rightConn, err: %s", err.Error())
	}
	bytespool.Free(buf)
	if rightConn != nil {
		if err := rightConn.Close(); err != nil {
			log.Errorf("Failed to close right conn, error: %s", err.Error())
		}
	}
	if leftConn != nil {
		if err := leftConn.Close(); err != nil {
			log.Errorf("Failed to close left conn, error: %s", err.Error())
		}
	}
}
