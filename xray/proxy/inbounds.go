package proxy

type Inbounds struct {
	Listen         string      `json:"listen,default=127.0.0.1"`
	Port           uint32      `json:"port,default=1080"`
	Protocol       string      `json:"protocol,default=socks"`
	Settings       interface{} `json:"settings"`
	StreamSettings interface{} `json:"streamSettings,optional"`
}
