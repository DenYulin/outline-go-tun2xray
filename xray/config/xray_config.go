package config

import (
	"github.com/DenYulin/outline-go-tun2xray/utils"
	"github.com/DenYulin/outline-go-tun2xray/xray/proxy"
)

type XrayConfig struct {
	LogConfig *LogConfig         `json:"logConfig"`
	Inbounds  []*proxy.Inbounds  `json:"inbounds"`
	Outbounds []*proxy.Outbounds `json:"outbounds"`
}

func (config *XrayConfig) String() string {
	return utils.ToJsonString(config)
}

type LogConfig struct {
	Access   string `json:"access,optional"`
	Error    string `json:"error,optional"`
	LogLevel string `json:"logLevel,default=info"`
	DnsLog   bool   `json:"dnsLog,default=false"`
}
