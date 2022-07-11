package xray

import (
	"context"
	"errors"
	"fmt"
	"github.com/eycorsican/go-tun2socks/common/log"
	t2core "github.com/eycorsican/go-tun2socks/core"
	"github.com/xtls/xray-core/common/bytespool"
	"github.com/xtls/xray-core/common/session"
	"github.com/xtls/xray-core/common/signal"
	"github.com/xtls/xray-core/common/task"
	"github.com/xxf098/go-tun2socks-build/pool"
	"net"
	"sync"
	"time"

	"github.com/xtls/xray-core/core"
)

type udpConnEntry struct {
	conn    net.PacketConn
	target  *net.UDPAddr
	updater signal.ActivityUpdater
}

type udpHandler struct {
	sync.Mutex

	ctx     context.Context
	v       *core.Instance
	connMap map[t2core.UDPConn]*udpConnEntry
	timeout time.Duration
}

func NewUDPHandler(ctx context.Context, instance *core.Instance, timeout time.Duration) t2core.UDPConnHandler {
	return &udpHandler{
		ctx:     ctx,
		v:       instance,
		connMap: make(map[t2core.UDPConn]*udpConnEntry, 16),
		timeout: timeout,
	}
}

func (handler *udpHandler) Connect(conn t2core.UDPConn, target *net.UDPAddr) error {
	if target == nil {
		return errors.New("nil target is not allowed")
	}
	sid := session.NewID()
	ctx := session.ContextWithID(handler.ctx, sid)
	ctx, cancel := context.WithCancel(ctx)
	pc, err := core.DialUDP(ctx, handler.v)
	if err != nil {
		cancel()
		return fmt.Errorf("dial V proxy connection failed: %v", err)
	}
	timer := signal.CancelAfterInactivity(ctx, cancel, handler.timeout)
	handler.Lock()
	handler.connMap[conn] = &udpConnEntry{
		conn:    pc,
		target:  target,
		updater: timer,
	}
	handler.Unlock()
	fetchTask := func() error {
		handler.fetchInput(conn)
		return nil
	}
	go func() {
		if err := task.Run(ctx, fetchTask); err != nil {
			pc.Close()
		}
	}()
	log.Infof("new proxy connection for target: %s:%s", target.Network(), target.String())
	return nil
}

func (handler *udpHandler) ReceiveTo(conn t2core.UDPConn, data []byte, addr *net.UDPAddr) error {
	handler.Lock()
	c, ok := handler.connMap[conn]
	handler.Unlock()

	if ok {
		_, err := c.conn.WriteTo(data, addr)
		c.updater.Update()
		if err != nil {
			handler.Close(conn)
			return fmt.Errorf("write remote failed: %v", err)
		}
		return nil
	} else {
		handler.Close(conn)
		return fmt.Errorf("proxy connection %v->%v does not exists", conn.LocalAddr(), addr)
	}
}

func (handler *udpHandler) Close(conn t2core.UDPConn) {
	handler.Lock()
	defer handler.Unlock()

	if c, found := handler.connMap[conn]; found {
		c.conn.Close()
	}
	delete(handler.connMap, conn)
}

func (handler *udpHandler) fetchInput(conn t2core.UDPConn) {
	handler.Lock()
	c, ok := handler.connMap[conn]
	handler.Unlock()
	if !ok {
		return
	}

	buf := bytespool.Alloc(pool.BufSize)
	defer bytespool.Free(buf)

	for {
		n, _, err := c.conn.ReadFrom(buf)
		if err != nil && n <= 0 {
			handler.Close(conn)
			conn.Close()
			return
		}
		c.updater.Update()
		_, err = conn.WriteFrom(buf[:n], c.target)
		if err != nil {
			handler.Close(conn)
			conn.Close()
			return
		}
	}
}
