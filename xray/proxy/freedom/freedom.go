package freedom

import (
	"encoding/json"
	"fmt"
	"github.com/DenYulin/outline-go-tun2xray/xray/proxy"
	"github.com/DenYulin/outline-go-tun2xray/xray/proxy/base"
	"github.com/eycorsican/go-tun2socks/common/log"
	"github.com/xtls/xray-core/infra/conf"
)

const (
	Tag      = "direct"
	Protocol = "freedom"
)

const (
	AsIs    = "AsIs"
	UseIP   = "UseIP"
	UseIPv4 = "UseIPv4"
	UseIPv6 = "UseIPv6"
)

type OutboundsSettings struct {
	DomainStrategy string `json:"domainStrategy,optional"`
	Redirect       string `json:"redirect,optional"`
}

func CreateFreedomOutboundDetourConfig(outbounds *proxy.Outbounds) (conf.OutboundDetourConfig, error) {
	var outboundDetourConfig conf.OutboundDetourConfig

	if outbounds.Protocol != Protocol {
		log.Errorf("The protocol must be freedom, current protocol: %s", outbounds.Protocol)
		return outboundDetourConfig, fmt.Errorf("the protocol must be freedom, current protocol: %s", outbounds.Protocol)
	}

	settings := outbounds.Settings.(OutboundsSettings)
	outboundsSettings, _ := json.Marshal(settings)
	outboundsSettingsMsg := json.RawMessage(outboundsSettings)

	outboundDetourConfig = conf.OutboundDetourConfig{
		Protocol:    Protocol,
		SendThrough: base.CreateSendThrough(outbounds.SendThrough),
		Tag:         Tag,
		Settings:    &outboundsSettingsMsg,
	}

	return outboundDetourConfig, nil
}

func CreateStreamSettings() *conf.StreamConfig {
	network := base.TCP
	security := base.NONE
	netProtocol := conf.TransportProtocol(network)
	streamSettings := &conf.StreamConfig{
		Network:  &netProtocol,
		Security: security,
	}
	return streamSettings
}
