package xray

import (
	"context"
	"github.com/DenYulin/outline-go-tun2xray/outline/common"
	"github.com/DenYulin/outline-go-tun2xray/tun2xray"
	"github.com/DenYulin/outline-go-tun2xray/xray"
	"github.com/xtls/xray-core/common/session"
	"net"
	"strconv"
)

func CheckConnectivity(serverAddress string, serverPort uint32, userId string) (int, error) {
	profile := &tun2xray.VLess{
		Host: serverAddress,
		Port: serverPort,
		ID:   userId,
	}

	xrayClient, err := tun2xray.StartXRayInstanceWithVLess(profile)
	if err != nil {
		return common.Unexpected, err
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
		return common.NoError, nil
	}

	_, isReachabilityError := tcpErr.(*xray.ReachabilityError)
	_, isAuthError := tcpErr.(*xray.AuthenticationError)
	if isAuthError {
		return common.AuthenticationFailure, nil
	} else if isReachabilityError {
		return common.Unreachable, nil
	}

	return common.Unexpected, tcpErr
}

// CheckServerReachable determines whether the server at `host:port` is reachable over TCP.
// Returns an error if the server is unreachable.
func CheckServerReachable(host string, port int) error {
	conn, err := net.DialTimeout("tcp", net.JoinHostPort(host, strconv.Itoa(port)), common.ReachabilityTimeout)
	if err != nil {
		return err
	}
	conn.Close()
	return nil
}
