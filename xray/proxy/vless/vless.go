package vless

import (
	"encoding/json"
	"fmt"
	"github.com/DenYulin/outline-go-tun2xray/xray/proxy"
	"github.com/google/martian/log"
	"github.com/xtls/xray-core/common/net"
	"github.com/xtls/xray-core/infra/conf"
)

const (
	Tag      = "vless-out"
	Protocol = "vless"
)

type Users struct {
	ID         string `json:"id"`
	Encryption int    `json:"encryption"`
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

func CreateVLessOutboundDetourConfig(outbounds proxy.Outbounds) (conf.OutboundDetourConfig, error) {
	var outboundDetourConfig conf.OutboundDetourConfig

	if outbounds.Protocol == Protocol {
		log.Errorf("The protocol must be socks, current protocol: %s", outbounds.Protocol)
		return outboundDetourConfig, fmt.Errorf("the protocol must be socks, current protocol: %s", outbounds.Protocol)
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

func CreateStreamSettings(network, security string) *conf.StreamConfig {
	netProtocol := conf.TransportProtocol(network)
	streamSettings := conf.StreamConfig{
		Network:  &netProtocol,
		Security: security,
	}

	switch network {
	case "tcp":
	case "kcp":
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
