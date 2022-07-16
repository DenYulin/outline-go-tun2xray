package xray

import (
	"context"
	"errors"
	"fmt"
	"github.com/DenYulin/outline-go-tun2xray/features"
	"github.com/DenYulin/outline-go-tun2xray/outline"
	"github.com/DenYulin/outline-go-tun2xray/tun2xray"
	"github.com/DenYulin/outline-go-tun2xray/tunnel"
	"github.com/DenYulin/outline-go-tun2xray/xray"
	t2core "github.com/eycorsican/go-tun2socks/core"
	"github.com/xtls/xray-core/common/session"
	x2core "github.com/xtls/xray-core/core"
	"io"
	"os"
	"time"
)

// Profile is the basic parameter used by xray startup
type Profile struct {
	Host             string
	Path             string
	InboundSocksPort uint32
	TLS              string
	Address          string
	Port             uint32
	Net              string
	ID               string
	Flow             string
	Type             string // headerType
	OutboundProtocol string `json:"protocol"`
	UseIPv6          bool   `json:"useIPv6"`
	LogLevel         string `json:"logLevel"`
	RouteMode        int    `json:"routeMode"`
	DNS              string `json:"DNS"`
	AllowInsecure    bool   `json:"allowInsecure"`
	Mux              int    `json:"mux"`
	AssetPath        string `json:"assetPath"`
}

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
		os.Setenv("xray.location.asset", profile.AssetPath)
		vLessProfile := toVLessProfile(profile)
		return tun2xray.StartXRayInstanceWithVLess(vLessProfile)
	default:
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
