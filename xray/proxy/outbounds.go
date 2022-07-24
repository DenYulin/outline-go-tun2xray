package proxy

import "github.com/DenYulin/outline-go-tun2xray/xray/proxy/base"

type Outbounds struct {
	SendThrough    string              `json:"sendThrough,default:0.0.0.0"`
	Protocol       string              `json:"protocol,default=vless"`
	Settings       interface{}         `json:"settings,optional"`
	StreamSettings base.StreamSettings `json:"streamSettings,optional"`
	ProxyTag       string              `json:"proxyTag,optional"`
	MuxConcurrency int16               `json:"mux,default=0"`
}
