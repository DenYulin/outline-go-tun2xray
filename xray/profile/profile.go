package profile

import (
	"github.com/DenYulin/outline-go-tun2xray/utils"
	"github.com/DenYulin/outline-go-tun2xray/xray/config"
	"github.com/DenYulin/outline-go-tun2xray/xray/proxy"
	"github.com/DenYulin/outline-go-tun2xray/xray/proxy/base"
	"github.com/DenYulin/outline-go-tun2xray/xray/proxy/freedom"
	"github.com/DenYulin/outline-go-tun2xray/xray/proxy/socks"
	"github.com/DenYulin/outline-go-tun2xray/xray/proxy/vless"
)

type Profile struct {
	LogLevel  string `json:"logLevel,default=info"`
	AccessLog string `json:"accessLog,optional"`
	ErrorLog  string `json:"errorLog,optional"`

	InboundListen   string `json:"inboundListen,default=127.0.0.1"`
	InboundProtocol string `json:"inboundProtocol,default=socks"`
	InboundPort     uint32 `json:"inboundPort,default=10800"`

	ServerAddress  string `json:"serverAddress"`
	ServerPort     uint32 `json:"serverPort,default=443"`
	Network        string `json:"network,default=tcp"`
	UserID         string `json:"id"`
	Flow           string `json:"flow,default=xtls-rprx-direct"`
	HeaderType     string `json:"headerType,default=none"`
	MuxConcurrency int16  `json:"mux,default=0"`
}

func (profile *Profile) String() string {
	return utils.ToJsonString(profile)
}

func CreateProfile(serverAddress string, serverPort uint32, userId string) *Profile {
	return &Profile{
		LogLevel:  base.DEBUG,
		AccessLog: "/outline/log/access.log",
		ErrorLog:  "/outline/log/error.log",

		InboundListen:   base.DefaultListenHost,
		InboundProtocol: socks.Protocol,
		InboundPort:     base.DefaultListenPort,

		ServerAddress:  serverAddress,
		ServerPort:     serverPort,
		Network:        base.DefaultNetwork,
		UserID:         userId,
		Flow:           vless.FlowXTlsRPrxDirect,
		HeaderType:     base.NONE,
		MuxConcurrency: 0,
	}
}

func ToXrayConfig(profile *Profile) *config.XrayConfig {
	logConfig := &config.LogConfig{
		Access:   profile.AccessLog,
		Error:    profile.ErrorLog,
		LogLevel: profile.LogLevel,
		DnsLog:   true,
	}

	inbounds := []*proxy.Inbounds{
		{
			Listen:   profile.InboundListen,
			Port:     profile.InboundPort,
			Protocol: socks.Protocol,
			Settings: socks.InboundsSettings{
				Auth: socks.DefaultAuth,
				IP:   base.DefaultListenHost,
				UDP:  base.DefaultEnableUDP,
			},
		},
	}

	outbounds := []*proxy.Outbounds{
		{
			SendThrough: base.DefaultSendThroughIp,
			Protocol:    vless.Protocol,
			Settings: vless.OutboundsSettings{
				Vnext: []vless.Vnext{
					{
						Address: profile.ServerAddress,
						Port:    profile.ServerPort,
						Users: []vless.Users{
							{
								ID:         profile.UserID,
								Encryption: base.DefaultEncryption,
								Flow:       vless.FlowXTlsRPrxDirect,
							},
						},
					},
				},
			},
			StreamSettings: base.StreamSettings{
				NetWork:    profile.Network,
				Security:   base.XTLS,
				HeaderType: profile.HeaderType,
				ServerName: profile.ServerAddress,
			},
			MuxConcurrency: base.DisableMux,
		},
		{
			SendThrough: base.DefaultSendThroughIp,
			Protocol:    freedom.Protocol,
			Settings: freedom.OutboundsSettings{
				DomainStrategy: freedom.UseIPv4,
			},
		},
	}

	xrayConfig := &config.XrayConfig{
		LogConfig: logConfig,
		Inbounds:  inbounds,
		Outbounds: outbounds,
	}

	return xrayConfig
}
