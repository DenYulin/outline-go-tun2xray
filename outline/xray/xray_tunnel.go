package xray

import (
	"context"
	"errors"
	"fmt"
	"github.com/DenYulin/outline-go-tun2xray/features"
	"github.com/DenYulin/outline-go-tun2xray/outline"
	"github.com/DenYulin/outline-go-tun2xray/tunnel"
	"github.com/DenYulin/outline-go-tun2xray/xray"
	"github.com/DenYulin/outline-go-tun2xray/xray/tun2xray"
	"github.com/eycorsican/go-tun2socks/common/log"
	t2core "github.com/eycorsican/go-tun2socks/core"
	"github.com/xtls/xray-core/common/session"
	x2core "github.com/xtls/xray-core/core"
	"io"
	"os"
	"time"
)

type outlineXrayTunnel struct {
	tunnel.Tunnel
	lwipStack  t2core.LWIPStack
	profile    *Profile
	xrayClient *x2core.Instance
}

func NewXrayTunnel(profile *Profile, tunWriter io.WriteCloser) (outline.Tunnel, error) {
	if tunWriter == nil {
		return nil, errors.New("must provide a TUN writer")
	}

	xrayClient, xrayErr := CreateXrayClient(profile)
	if xrayErr != nil {
		return nil, fmt.Errorf("invalid xray proxy parameters, error: %s", xrayErr.Error())
	}

	lwipStack := t2core.NewLWIPStack()
	tunnel := tunnel.NewTunnel(tunWriter, lwipStack)
	xrayTunnel := &outlineXrayTunnel{Tunnel: tunnel, lwipStack: lwipStack, profile: profile, xrayClient: xrayClient}
	xrayTunnel.registerConnectionHandlers()
	return xrayTunnel, nil
}

func NewXrayTunnelWithJson(configJson string, tunWriter io.WriteCloser) (outline.Tunnel, error) {
	if tunWriter == nil {
		return nil, errors.New("must provide a TUN writer")
	}

	xrayClient, xrayErr := tun2xray.StartXRayInstanceWithJson(configJson)
	if xrayErr != nil {
		return nil, fmt.Errorf("invalid xray proxy parameters, error: %s", xrayErr.Error())
	}

	lwipStack := t2core.NewLWIPStack()
	tunnel := tunnel.NewTunnel(tunWriter, lwipStack)
	xrayTunnel := &outlineXrayTunnel{Tunnel: tunnel, lwipStack: lwipStack, profile: nil, xrayClient: xrayClient}
	xrayTunnel.registerConnectionHandlers()
	return xrayTunnel, nil
}

func (t *outlineXrayTunnel) UpdateUDPSupport() bool {
	return true
}

func (t *outlineXrayTunnel) registerConnectionHandlers() {
	ctx := context.Background()
	content := session.ContentFromContext(ctx)
	if content == nil {
		content = new(session.Content)
		ctx = session.ContextWithContent(ctx, content)
	}

	t2core.RegisterTCPConnHandler(xray.NewTCPHandler(ctx, t.xrayClient))
	t2core.RegisterUDPConnHandler(xray.NewUDPHandler(ctx, t.xrayClient, 3*time.Minute))
}

func CreateXrayClient(profile *Profile) (*x2core.Instance, error) {
	if profile == nil {
		return nil, errors.New("create xray client error")
	}

	switch profile.OutboundProtocol {
	case tun2xray.VLESS:
		log.Infof("Start to create xray client with VLess protocol")
		key := "xray.location.asset"
		if err := os.Setenv(key, profile.AssetPath); err != nil {
			log.Errorf("Failed to set a env param, key: %s, value: %s, error: %+v", key, profile.AssetPath, err)
			return nil, err
		}
		vLessProfile := toVLessProfile(profile)
		return tun2xray.StartXRayInstanceWithVLess(vLessProfile)
	default:
		log.Warnf("unsupported xray outbound protocol, protocol: %s", profile.OutboundProtocol)
		return nil, errors.New("unsupported xray outbound protocol")
	}
}

func toVLessProfile(profile *Profile) *tun2xray.VLess {
	vLessProfile := &tun2xray.VLess{
		Host:     profile.Host,
		Path:     profile.Path,
		TLS:      profile.TLS,
		Address:  profile.Address,
		Port:     profile.Port,
		Net:      profile.Net,
		ID:       profile.ID,
		Flow:     profile.Flow,
		Type:     profile.Type,
		Protocol: profile.OutboundProtocol,
		VLessOptions: features.VLessOptions{
			UseIPv6:       profile.UseIPv6,
			LogLevel:      profile.LogLevel,
			RouteMode:     profile.RouteMode,
			DNS:           profile.DNS,
			AllowInsecure: profile.AllowInsecure,
			Mux:           profile.Mux,
		},
	}
	return vLessProfile
}
