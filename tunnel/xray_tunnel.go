package tunnel

import (
	"context"
	"errors"
	"github.com/DenYulin/outline-go-tun2xray/xray"
	"github.com/DenYulin/outline-go-tun2xray/xray/profile"
	"github.com/DenYulin/outline-go-tun2xray/xray/tun2xray"
	"github.com/eycorsican/go-tun2socks/common/log"
	t2core "github.com/eycorsican/go-tun2socks/core"
	"github.com/xtls/xray-core/common/session"
	x2core "github.com/xtls/xray-core/core"
	"io"
	"time"
)

type OutlineXrayTunnel struct {
	Tunnel
	lwipStack  t2core.LWIPStack
	profile    *profile.Profile
	xrayClient *x2core.Instance
}

func NewXrayTunnel(profile *profile.Profile, tunWriter io.WriteCloser) (*OutlineXrayTunnel, error) {
	if tunWriter == nil {
		log.Errorf("Must provide a TUN writer")
		return nil, errors.New("must provide a TUN writer")
	}

	xrayClient, err := tun2xray.StartXRayInstanceWithVLessAndXTLS(profile)
	if err != nil {
		log.Errorf("Failed to start xray client with vless and xtls, profile: %s, error: %s", profile.String(), err.Error())
		return nil, err
	}

	lwipStack := t2core.NewLWIPStack()
	tunnel := NewTunnel(tunWriter, lwipStack)
	xrayTunnel := &OutlineXrayTunnel{Tunnel: tunnel, lwipStack: lwipStack, profile: profile, xrayClient: xrayClient}
	xrayTunnel.registerConnectionHandlers()
	return xrayTunnel, nil
}

func (t *OutlineXrayTunnel) UpdateUDPSupport() bool {
	return true
}

func (t *OutlineXrayTunnel) registerConnectionHandlers() {
	ctx := context.Background()
	content := session.ContentFromContext(ctx)
	if content == nil {
		content = new(session.Content)
		ctx = session.ContextWithContent(ctx, content)
	}

	t2core.RegisterTCPConnHandler(xray.NewTCPHandler(ctx, t.xrayClient))
	t2core.RegisterUDPConnHandler(xray.NewUDPHandler(ctx, t.xrayClient, 3*time.Minute))
}
