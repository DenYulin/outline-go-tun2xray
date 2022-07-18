package xray

import (
	"context"
	"github.com/DenYulin/outline-go-tun2xray/xray"
	"github.com/DenYulin/outline-go-tun2xray/xray/tun2xray"
	"github.com/eycorsican/go-tun2socks/common/log"
	"github.com/xtls/xray-core/common/session"
	"net"
	"strconv"
)

func CheckConnectivity(serverAddress string, serverPort int, userId string) (int, error) {
	profile := &tun2xray.VLess{
		Host: serverAddress,
		Port: uint32(serverPort),
		ID:   userId,
	}

	xrayClient, err := tun2xray.StartXRayInstanceWithVLess(profile)
	if err != nil {
		return Unexpected, err
	}

	ctx := context.Background()
	content := session.ContentFromContext(ctx)
	if content == nil {
		content = new(session.Content)
		ctx = session.ContextWithContent(ctx, content)
	}

	tcpChan := make(chan error)
	go func() {
		tcpChan <- xray.CheckTCPConnectivityWithHTTP(ctx, xrayClient, "http://example.com")
	}()

	tcpErr := <-tcpChan
	if tcpErr == nil {
		// The TCP connectivity checks succeeded, which means UDP is not supported.
		return NoError, nil
	}

	_, isReachabilityError := tcpErr.(*xray.ReachabilityError)
	_, isAuthError := tcpErr.(*xray.AuthenticationError)
	if isAuthError {
		return AuthenticationFailure, nil
	} else if isReachabilityError {
		return Unreachable, nil
	}

	return Unexpected, tcpErr
}

// CheckServerReachable determines whether the server at `host:port` is reachable over TCP.
// Returns an error if the server is unreachable.
func CheckServerReachable(host string, port int) error {
	conn, err := net.DialTimeout("tcp", net.JoinHostPort(host, strconv.Itoa(port)), ReachabilityTimeout)
	if err != nil {
		log.Errorf("Failed to tcp dial, host: %s, port: %d, error: %+v", host, port, err)
		return err
	}
	if err := conn.Close(); err != nil {
		log.Errorf("Failed to close tcp connect, error: %+v", err)
		return err
	}
	return nil
}
