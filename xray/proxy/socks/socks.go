package socks

import (
	"encoding/json"
	"fmt"
	"github.com/DenYulin/outline-go-tun2xray/xray/proxy"
	"github.com/eycorsican/go-tun2socks/common/log"
	"github.com/xtls/xray-core/common/net"
	"github.com/xtls/xray-core/infra/conf"
)

const (
	Tag      = "socks-in"
	Protocol = "socks"
)

const DefaultAuth = "noauth"

type InboundsSettings struct {
	Auth      string `json:"auth,default=noauth"`
	IP        string `json:"ip,default=127.0.0.1"`
	UDP       bool   `json:"udp,default=true"`
	UserLevel int32  `json:"userLevel,optional"`
}

func CreateSocks5InboundDetourConfig(inbounds *proxy.Inbounds) (conf.InboundDetourConfig, error) {
	var inboundDetourConfig conf.InboundDetourConfig

	if inbounds.Protocol != Protocol {
		log.Errorf("The protocol must be socks, current protocol: %s", inbounds.Protocol)
		return inboundDetourConfig, fmt.Errorf("the protocol must be socks, current protocol: %s", inbounds.Protocol)
	}

	settings := inbounds.Settings.(InboundsSettings)
	inboundsSettings, _ := json.Marshal(InboundsSettings{
		Auth: settings.Auth,
		IP:   settings.IP,
		UDP:  settings.UDP,
	})

	inboundsSettingsMsg := json.RawMessage(inboundsSettings)
	inboundDetourConfig = conf.InboundDetourConfig{
		Tag:      Tag,
		Protocol: Protocol,
		PortList: &conf.PortList{Range: []conf.PortRange{{From: inbounds.Port, To: inbounds.Port}}},
		ListenOn: &conf.Address{Address: net.IPAddress([]byte{127, 0, 0, 1})},
		Settings: &inboundsSettingsMsg,
	}

	return inboundDetourConfig, nil
}
