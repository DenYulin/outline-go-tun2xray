package socks

import (
	"encoding/json"
	"fmt"
	"github.com/DenYulin/outline-go-tun2xray/xray/proxy"
	"github.com/xtls/xray-core/common/net"
	"github.com/xtls/xray-core/infra/conf"
)

const (
	Tag      = "socks-in"
	Protocol = "socks"
)

type InboundsSettings struct {
	Auth      string `json:"auth,default="`
	IP        string `json:"ip"`
	UDP       bool   `json:"udp"`
	UserLevel int32  `json:"userLevel"`
}

func CreateSocks5InboundDetourConfig(inbounds *proxy.Inbounds) (conf.InboundDetourConfig, error) {
	var inboundsDetourConfig conf.InboundDetourConfig

	if inbounds.Protocol != Protocol {
		return inboundsDetourConfig, fmt.Errorf("the protocol must be socks, param protocol: %s", inbounds.Protocol)
	}

	settings := inbounds.Settings.(InboundsSettings)
	inboundsSettings, _ := json.Marshal(InboundsSettings{
		Auth: settings.Auth,
		IP:   settings.IP,
		UDP:  settings.UDP,
	})

	inboundsSettingsMsg := json.RawMessage(inboundsSettings)
	inboundsDetourConfig = conf.InboundDetourConfig{
		Tag:      Tag,
		Protocol: Protocol,
		PortList: &conf.PortList{Range: []conf.PortRange{{From: inbounds.Port, To: inbounds.Port}}},
		ListenOn: &conf.Address{Address: net.IPAddress([]byte{127, 0, 0, 1})},
		Settings: &inboundsSettingsMsg,
	}

	return inboundsDetourConfig, nil
}
