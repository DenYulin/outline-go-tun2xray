package proxy

import "github.com/DenYulin/outline-go-tun2xray/xray/proxy/base"

type Inbounds struct {
	Listen         string              `json:"listen,default=127.0.0.1"`
	Port           uint32              `json:"port,default=1080"`
	Protocol       string              `json:"protocol,default=socks"`
	Settings       interface{}         `json:"settings"`
	StreamSettings base.StreamSettings `json:"streamSettings,optional"`
}
