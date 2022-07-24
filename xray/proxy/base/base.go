package base

import (
	"encoding/json"
	"github.com/xtls/xray-core/common/net"
	"github.com/xtls/xray-core/infra/conf"
)

const (
	TCP          = "tcp"
	KCP          = "kcp"
	WS           = "ws"
	HTTP         = "http"
	DomainSocket = "domainsocket"
	QUIC         = "quic"
	GRPC         = "grpc"
)

const (
	NONE = "none"
	TLS  = "tls"
	XTLS = "xtls"
)

const (
	DefaultListenHost = "127.0.0.1"
	DefaultListenPort = 10800
	DefaultEncryption = "none"
	DefaultEnableUDP  = true
	DefaultNetwork    = "tcp"
)

const (
	DefaultSendThroughIp = "0.0.0.0"
	DisableMux           = 0
)

const (
	DEBUG = "debug"
	INFO  = "info"
)

type StreamSettings struct {
	NetWork    string `json:"netWork"`
	Security   string `json:"security"`
	HeaderType string `json:"headerType"`
	ServerName string `json:"serverName"`
}

type HeaderConfig struct {
	Type string `json:"type"`
}

func CreateHeaderConfig(headerType string) json.RawMessage {
	headerConfig := HeaderConfig{
		Type: headerType,
	}

	config, _ := json.Marshal(headerConfig)
	configMsg := json.RawMessage(config)

	return configMsg
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
