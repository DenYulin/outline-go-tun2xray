package proxy

type Outbounds struct {
	SendThrough    string      `json:"sendThrough,default:0.0.0.0"`
	Protocol       string      `json:"protocol,default=vless"`
	Settings       interface{} `json:"settings,optional"`
	Tag            string      `json:"tag"`
	StreamSettings interface{} `json:"streamSettings,optional"`
	ProxyTag       string      `json:"proxyTag,optional"`
	MuxConcurrency int16       `json:"mux,default=0"`
}
