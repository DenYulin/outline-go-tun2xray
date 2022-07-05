package features

import (
	"github.com/Jigsaw-Code/outline-go-tun2socks/common"
	"github.com/xtls/xray-core/infra/conf"
)

type VLessOptions struct {
	LogLevel  string `json:"logLevel"`
	RouteMode int    `json:"routeMode"`
	DNS       string `json:"DNS"`
}

type VLess struct {
	Protocol string `json:"protocol"`
	VLessOptions
}

func LoadVLessConfig(profile *VLess) {
	jsonConfig := &conf.Config{}
	jsonConfig.LogConfig = &conf.LogConfig{
		LogLevel: profile.LogLevel,
	}

	// https://github.com/Loyalsoldier/v2ray-rules-dat
	jsonConfig.DNSConfig = common.CreateDNSConfig(profile.RouteMode, profile.DNS)

	// update rules
	jsonConfig.RouterConfig = common.CreateRouterConfig(profile.RouteMode)
}

func (profile *VLess) getProxyOutboundDetourConfig() conf.OutboundDetourConfig {
	proxyOutboundConfig := conf.OutboundDetourConfig{}
	//if profile.Protocol == common.VMess {
	//    proxyOutboundConfig = createVmessOutboundDetourConfig(profile)
	//}
	//if profile.Protocol == common.Trojan {
	//    proxyOutboundConfig = createTrojanOutboundDetourConfig(profile)
	//}
	if profile.Protocol == common.VLess {
		proxyOutboundConfig = common.CreateVLessOutboundDetourConfig(profile)
	}
	return proxyOutboundConfig
}
