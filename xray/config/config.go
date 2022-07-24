package config

import (
	"github.com/DenYulin/outline-go-tun2xray/xray/proxy"
	"github.com/DenYulin/outline-go-tun2xray/xray/proxy/freedom"
	"github.com/DenYulin/outline-go-tun2xray/xray/proxy/socks"
	"github.com/DenYulin/outline-go-tun2xray/xray/proxy/vless"
	"github.com/xtls/xray-core/infra/conf"
)

func LoadXrayConfig(xrayConfig *XrayConfig) (*conf.Config, error) {
	jsonConfig := &conf.Config{}
	jsonConfig.LogConfig = CreateLogConfig(xrayConfig.LogConfig)

	jsonConfig.InboundConfigs = CreateInboundConfigs(xrayConfig.Inbounds)
	jsonConfig.OutboundConfigs = CreateOutboundConfigs(xrayConfig.Outbounds)

	return jsonConfig, nil
}

func CreateLogConfig(logConfig *LogConfig) *conf.LogConfig {
	return &conf.LogConfig{
		LogLevel:  logConfig.LogLevel,
		AccessLog: logConfig.Access,
		ErrorLog:  logConfig.Error,
	}
}

func CreateInboundConfigs(inbounds []*proxy.Inbounds) []conf.InboundDetourConfig {
	inboundDetourConfigs := make([]conf.InboundDetourConfig, 0)
	for _, inbound := range inbounds {
		switch inbound.Protocol {
		case socks.Protocol:
			socks5InboundDetourConfig, _ := socks.CreateSocks5InboundDetourConfig(inbound)
			inboundDetourConfigs = append(inboundDetourConfigs, socks5InboundDetourConfig)
		}
	}

	return inboundDetourConfigs
}

func CreateOutboundConfigs(outbounds []*proxy.Outbounds) []conf.OutboundDetourConfig {
	outboundDetourConfigs := make([]conf.OutboundDetourConfig, 0)
	for _, outbound := range outbounds {
		switch outbound.Protocol {
		case vless.Protocol:
			vLessOutboundDetourConfig, _ := vless.CreateVLessOutboundDetourConfig(outbound)
			outboundDetourConfigs = append(outboundDetourConfigs, vLessOutboundDetourConfig)
		case freedom.Protocol:
			freedomOutboundDetourConfig, _ := freedom.CreateFreedomOutboundDetourConfig(outbound)
			outboundDetourConfigs = append(outboundDetourConfigs, freedomOutboundDetourConfig)
		}
	}

	return outboundDetourConfigs
}
