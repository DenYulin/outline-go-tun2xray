package vless

import (
	"encoding/json"
	"fmt"
	"github.com/DenYulin/outline-go-tun2xray/xray/proxy"
	"github.com/DenYulin/outline-go-tun2xray/xray/proxy/base"
	"github.com/eycorsican/go-tun2socks/common/log"
	"github.com/xtls/xray-core/common/net"
	"github.com/xtls/xray-core/infra/conf"
)

const (
	Tag      = "vless-out"
	Protocol = "vless"
)

const (
	FlowXTlsRPrxDirect = "xtls-rprx-direct"
)

type Users struct {
	ID         string `json:"id"`
	Encryption string `json:"encryption"`
	Flow       string `json:"flow"`
	Level      int    `json:"level"`
}

type Vnext struct {
	Address string  `json:"address"`
	Port    uint32  `json:"port"`
	Users   []Users `json:"users"`
}

type OutboundsSettings struct {
	Vnext []Vnext `json:"vnext,omitempty"`
}

func CreateVLessOutboundDetourConfig(outbounds *proxy.Outbounds) (conf.OutboundDetourConfig, error) {
	var outboundDetourConfig conf.OutboundDetourConfig

	if outbounds.Protocol != Protocol {
		log.Errorf("The protocol must be VLess, current protocol: %s", outbounds.Protocol)
		return outboundDetourConfig, fmt.Errorf("the protocol must be VLess, current protocol: %s", outbounds.Protocol)
	}

	settings := outbounds.Settings.(OutboundsSettings)
	outboundsSettings, _ := json.Marshal(settings)
	outboundsSettingsMsg := json.RawMessage(outboundsSettings)

	outboundDetourConfig = conf.OutboundDetourConfig{
		Protocol:      Protocol,
		SendThrough:   CreateSendThrough(outbounds.SendThrough),
		Tag:           Tag,
		Settings:      &outboundsSettingsMsg,
		ProxySettings: CreateProxySettings(outbounds.ProxyTag),
		StreamSetting: CreateStreamSettings(outbounds.StreamSettings),
		MuxSettings:   CreateMuxConfig(outbounds.MuxConcurrency),
	}

	return outboundDetourConfig, nil
}

func CreateSendThrough(sendThrough string) *conf.Address {
	return &conf.Address{
		Address: net.ParseAddress(sendThrough),
	}
}

func CreateProxySettings(proxyTag string) *conf.ProxyConfig {
	var proxySettings *conf.ProxyConfig
	if len(proxyTag) > 0 {
		proxySettings.Tag = proxyTag
	}
	return proxySettings
}

func CreateStreamSettings(settings base.StreamSettings) *conf.StreamConfig {
	network := settings.NetWork
	security := settings.Security
	netProtocol := conf.TransportProtocol(network)
	streamSettings := conf.StreamConfig{
		Network:  &netProtocol,
		Security: security,
	}

	switch network {
	case base.TLS:
		if security == base.TLS {
			tlsConfig := &conf.TLSConfig{
				ServerName: settings.ServerName,
			}
			streamSettings.TLSSettings = tlsConfig
		} else if security == base.XTLS {
			xtlsConfig := &conf.XTLSConfig{
				ServerName: settings.ServerName,
				ALPN:       &conf.StringList{"h2", "http/1.1"},
				MinVersion: "1.2",
				MaxVersion: "1.3",
			}
			streamSettings.XTLSSettings = xtlsConfig
		}
	case base.KCP:
		mtu := uint32(1350)
		tti := uint32(50)
		upCap := uint32(12)
		downCap := uint32(100)
		congestion := false
		readBufferSize := uint32(2)
		writeBufferSize := uint32(2)

		kcpSettings := &conf.KCPConfig{
			Mtu:             &mtu,
			Tti:             &tti,
			UpCap:           &upCap,
			DownCap:         &downCap,
			Congestion:      &congestion,
			ReadBufferSize:  &readBufferSize,
			WriteBufferSize: &writeBufferSize,
			HeaderConfig:    base.CreateHeaderConfig(settings.HeaderType),
		}
		streamSettings.KCPSettings = kcpSettings
	}

	return &streamSettings
}

func CreateMuxConfig(muxConcurrency int16) *conf.MuxConfig {
	enabled := false
	concurrency := muxConcurrency
	if muxConcurrency > 0 {
		enabled = true
	} else {
		concurrency = 0
	}

	return &conf.MuxConfig{
		Enabled:     enabled,
		Concurrency: concurrency,
	}
}
