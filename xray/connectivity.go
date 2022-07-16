package xray

import (
	"context"
	"github.com/eycorsican/go-tun2socks/common/log"
	x2net "github.com/xtls/xray-core/common/net"
	"github.com/xtls/xray-core/common/session"
	"github.com/xtls/xray-core/core"
	"net"
	"net/http"
	"time"
)

const (
	tcpTimeoutMs        = udpTimeoutMs * udpMaxRetryAttempts
	udpTimeoutMs        = 1000
	udpMaxRetryAttempts = 5
	bufferLength        = 512
)

const (
	HEAD = "HEAD"
	TCP  = "tcp"
)

// AuthenticationError is used to signal failed authentication to the xray proxy.
type AuthenticationError struct {
	error
}

// ReachabilityError is used to signal an unreachable proxy.
type ReachabilityError struct {
	error
}

func CheckTCPConnectivityWithHTTP(ctx context.Context, xrayClient *core.Instance, targetAddr string) error {
	request, err := http.NewRequest(HEAD, targetAddr, nil)
	if err != nil {
		log.Errorf("Create new http request error: %+v", err)
		return err
	}
	targetAddr = request.Host
	if !hasPort(targetAddr) {
		targetAddr = net.JoinHostPort(targetAddr, "80")
	}

	conn, err := net.Dial(TCP, targetAddr)
	if err != nil {
		log.Errorf("Connects to the target address error: %+v", err)
		return err
	}

	dest := x2net.DestinationFromAddr(conn.RemoteAddr())
	sid := session.NewID()
	sCtx := session.ContextWithID(ctx, sid)
	destConn, err := core.Dial(sCtx, xrayClient, dest)
	if err != nil {
		log.Errorf("Dial xray proxy connection failed, error: %+v", err)
		return err
	}
	defer func(conn net.Conn) {
		if err := conn.Close(); err != nil {
			log.Errorf("Close connect error: %+v", err)
		}
	}(conn)

	if err = destConn.SetDeadline(time.Now().Add(time.Millisecond * tcpTimeoutMs)); err != nil {
		log.Errorf("Set deadline to dest connect error: %+v", err)
		return err
	}
	err = request.Write(conn)
	if err != nil {
		return &AuthenticationError{err}
	}
	n, err := conn.Read(make([]byte, bufferLength))
	if n == 0 && err != nil {
		return &AuthenticationError{err}
	}

	return nil
}

func hasPort(hostPort string) bool {
	_, _, err := net.SplitHostPort(hostPort)
	return err == nil
}
