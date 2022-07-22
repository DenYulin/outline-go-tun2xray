package tun2xray

import (
	"encoding/json"
	"github.com/DenYulin/outline-go-tun2xray/xray"
	"github.com/DenYulin/outline-go-tun2xray/xray/proxy/dokodemoDoor"
	"github.com/eycorsican/go-tun2socks/common/log"
	_ "github.com/xtls/xray-core/common"
	xnet "github.com/xtls/xray-core/common/net"
	"github.com/xtls/xray-core/infra/conf"
	"net"
	"strconv"
	"strings"
)

const SeparatorComma = ","

const (
	DOKODEMO_DOOR        = "dokodemo-door"
	SOCKS         string = "socks"
	VLESS         string = "vless"
)

func CreateDNSConfig(routeMode int, dnsConf string) *conf.DNSConfig {
	dns := strings.Split(dnsConf, SeparatorComma)
	nameServerConfig := make([]*conf.NameServerConfig, 0)
	if routeMode == 2 || routeMode == 3 || routeMode == 4 {
		for i := len(dns) - 1; i >= 0; i-- {
			if newConfig := toNameServerConfig(dns[i]); newConfig != nil {
				if i == 1 {
					newConfig.Domains = []string{"geosite:cn"}
				}
				nameServerConfig = append(nameServerConfig, newConfig)
			}
		}
	} else {
		if newConfig := toNameServerConfig(dns[0]); newConfig != nil {
			nameServerConfig = append(nameServerConfig, newConfig)
		}
	}
	return &conf.DNSConfig{
		//Hosts:   xray.BlockHosts,
		Servers: nameServerConfig,
	}
}

func toNameServerConfig(hostPort string) *conf.NameServerConfig {
	host, port, err := net.SplitHostPort(hostPort)
	if err != nil {
		log.Errorf("Split host and port error, return nil, error: %s", err.Error())
		return nil
	}
	iPort, err := strconv.Atoi(port)
	if err != nil {
		log.Errorf("Convert the string to the integer error, return nil, source port: %s, error: %s", port, err.Error())
		return nil
	}
	newConfig := &conf.NameServerConfig{
		Address: &conf.Address{Address: xnet.ParseAddress(host)},
		Port:    uint16(iPort),
	}
	return newConfig
}

func CreateDokodemoDoorInboundDetourConfig(proxyPort uint32) conf.InboundDetourConfig {
	inboundsSettings, _ := json.Marshal(dokodemoDoor.InboundsSettings{
		Address:        "127.0.0.1",
		Port:           proxyPort,
		Network:        "tcp,udp",
		Timeout:        300,
		FollowRedirect: false,
		UserLevel:      0,
	})

	inboundsSettingsMsg := json.RawMessage(inboundsSettings)
	inboundsDetourConfig := conf.InboundDetourConfig{
		Tag:      "transparent",
		Protocol: "dokodemo-door",
		PortList: &conf.PortList{Range: []conf.PortRange{{From: proxyPort, To: proxyPort}}},
		ListenOn: &conf.Address{Address: xnet.IPAddress([]byte{127, 0, 0, 1})},
		Settings: &inboundsSettingsMsg,
	}

	return inboundsDetourConfig
}

func CreateSocks5InboundDetourConfig(proxyPort uint32) conf.InboundDetourConfig {
	inboundsSettings, _ := json.Marshal(xray.InboundsSettings{
		Auth: "noauth",
		IP:   "127.0.0.1",
		UDP:  true,
	})

	inboundsSettingsMsg := json.RawMessage(inboundsSettings)
	inboundsDetourConfig := conf.InboundDetourConfig{
		Tag:      "socks-in",
		Protocol: "socks",
		PortList: &conf.PortList{Range: []conf.PortRange{{From: proxyPort, To: proxyPort}}},
		ListenOn: &conf.Address{Address: xnet.IPAddress([]byte{127, 0, 0, 1})},
		Settings: &inboundsSettingsMsg,
	}

	return inboundsDetourConfig
}

func CreateVLessOutboundDetourConfig(profile *VLess) conf.OutboundDetourConfig {
	outboundsSettings, _ := json.Marshal(xray.OutboundsSettings{
		Vnext: []xray.Vnext{
			{
				Address: profile.ServerAddress,
				Port:    profile.ServerPort,
				Users: []xray.Users{
					{
						ID:         profile.ID,
						Encryption: "none",
						Flow:       profile.Flow,
						Level:      8,
					},
				},
			},
		},
	})

	outboundsSettingsMsg := json.RawMessage(outboundsSettings)
	muxEnabled := false
	if profile.Mux > 0 {
		muxEnabled = true
	} else {
		profile.Mux = -1
	}
	tcp := conf.TransportProtocol("tcp")
	vLessOutboundDetourConfig := conf.OutboundDetourConfig{
		Protocol: "vless",
		Tag:      "proxy",
		MuxSettings: &conf.MuxConfig{
			Enabled:     muxEnabled,
			Concurrency: int16(profile.Mux),
		},
		Settings: &outboundsSettingsMsg,
		StreamSetting: &conf.StreamConfig{
			Network:  &tcp,
			Security: "",
		},
	}

	if profile.Net == "ws" {
		transportProtocol := conf.TransportProtocol(profile.Net)
		vLessOutboundDetourConfig.StreamSetting = &conf.StreamConfig{
			Network:    &transportProtocol,
			WSSettings: &conf.WebSocketConfig{Path: profile.Path},
		}
		if profile.Host != "" {
			vLessOutboundDetourConfig.StreamSetting.WSSettings.Headers = map[string]string{"Host": profile.Host}
		}
	}

	if profile.Net == "h2" {
		transportProtocol := conf.TransportProtocol(profile.Net)
		vLessOutboundDetourConfig.StreamSetting = &conf.StreamConfig{
			Network:      &transportProtocol,
			HTTPSettings: &conf.HTTPConfig{Path: profile.Path},
		}
		if profile.Host != "" {
			hosts := strings.Split(profile.Host, ",")
			vLessOutboundDetourConfig.StreamSetting.HTTPSettings.Host = conf.NewStringList(hosts)
		}
	}

	if profile.Net == "quic" {
		transportProtocol := conf.TransportProtocol(profile.Net)
		vLessOutboundDetourConfig.StreamSetting = &conf.StreamConfig{
			Network:      &transportProtocol,
			QUICSettings: &conf.QUICConfig{Key: profile.Path},
		}
		if profile.Host != "" {
			vLessOutboundDetourConfig.StreamSetting.QUICSettings.Security = profile.Host
		}
		if profile.Type != "" {
			header, _ := json.Marshal(xray.QUICSettingsHeader{Type: profile.Type})
			vLessOutboundDetourConfig.StreamSetting.QUICSettings.Header = header
		}
	}

	if profile.Net == "kcp" {
		transportProtocol := conf.TransportProtocol(profile.Net)
		mtu := uint32(1350)
		tti := uint32(50)
		upCap := uint32(12)
		downCap := uint32(100)
		congestion := false
		readBufferSize := uint32(1)
		writeBufferSize := uint32(1)
		vLessOutboundDetourConfig.StreamSetting = &conf.StreamConfig{
			Network: &transportProtocol,
			KCPSettings: &conf.KCPConfig{
				Mtu:             &mtu,
				Tti:             &tti,
				UpCap:           &upCap,
				DownCap:         &downCap,
				Congestion:      &congestion,
				ReadBufferSize:  &readBufferSize,
				WriteBufferSize: &writeBufferSize,
			},
		}
		if profile.Type != "" {
			header, _ := json.Marshal(xray.KCPSettingsHeader{Type: profile.Type})
			vLessOutboundDetourConfig.StreamSetting.KCPSettings.HeaderConfig = json.RawMessage(header)
		}
	}

	// tcp 带 http 伪装
	if profile.Net == "tcp" && profile.Type == "http" {
		transportProtocol := conf.TransportProtocol(profile.Net)
		tcpSettingsHeader := xray.TCPSettingsHeader{
			Type: profile.Type,
			TCPSettingsRequest: xray.TCPSettingsRequest{
				Version: "1.1",
				Method:  "GET",
				Path:    strings.Split(profile.Path, ","),
				Headers: xray.HTTPHeaders{
					UserAgent:      []string{"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/80.0.3987.149 Safari/537.36"},
					AcceptEncoding: []string{"gzip, deflate"},
					Connection:     "keep-alive",
					Pragma:         "no-cache",
					Host:           strings.Split(profile.Host, ","),
				},
			},
		}
		header, _ := json.Marshal(tcpSettingsHeader)
		vLessOutboundDetourConfig.StreamSetting = &conf.StreamConfig{
			Network:  &transportProtocol,
			Security: profile.TLS,
			TCPSettings: &conf.TCPConfig{
				HeaderConfig: header,
			},
		}
	}

	if profile.TLS == "tls" {
		vLessOutboundDetourConfig.StreamSetting.Security = profile.TLS
		tlsConfig := &conf.TLSConfig{Insecure: profile.AllowInsecure}
		if profile.Host != "" {
			tlsConfig.ServerName = profile.Host
		}
		vLessOutboundDetourConfig.StreamSetting.TLSSettings = tlsConfig
	}

	if profile.TLS == "xtls" {
		vLessOutboundDetourConfig.StreamSetting.Security = profile.TLS
		xTlsConfig := &conf.XTLSConfig{Insecure: profile.AllowInsecure}
		if profile.Host != "" {
			xTlsConfig.ServerName = profile.Host
		}
		vLessOutboundDetourConfig.StreamSetting.XTLSSettings = xTlsConfig
	}

	return vLessOutboundDetourConfig
}

func CreatePolicyConfig() *conf.PolicyConfig {
	handshake := uint32(4)
	connIdle := uint32(300)
	downLinkOnly := uint32(1)
	uplinkOnly := uint32(1)
	return &conf.PolicyConfig{
		Levels: map[uint32]*conf.Policy{
			8: {
				ConnectionIdle: &connIdle,
				DownlinkOnly:   &downLinkOnly,
				Handshake:      &handshake,
				UplinkOnly:     &uplinkOnly,
			},
		},
		System: &conf.SystemPolicy{
			StatsOutboundUplink:   true,
			StatsOutboundDownlink: true,
		},
	}
}

func CreateFreedomOutboundDetourConfig(useIPv6 bool) conf.OutboundDetourConfig {
	domainStrategy := "UseIPv4"
	if useIPv6 {
		domainStrategy = "UseIP"
	}
	outboundsSettings, _ := json.Marshal(xray.OutboundsSettings{DomainStrategy: domainStrategy})
	outboundsSettingsMsg := json.RawMessage(outboundsSettings)
	return conf.OutboundDetourConfig{
		Protocol: "freedom",
		Tag:      "direct",
		Settings: &outboundsSettingsMsg,
	}
}

// CreateRouterConfig
// 0 all
// 1 bypass LAN
// 2 bypass China
// 3 bypass LAN & China
// 4 GFWList`
// 5 ChinaList
// >= 6 bypass LAN & China & AD block
// 	0: "Plain", 1: "Regex", 2: "Domain", 3: "Full",
//
// https://github.com/Loyalsoldier/v2ray-rules-dat
func CreateRouterConfig(routeMode int) *conf.RouterConfig {
	domainStrategy := "IPIfNonMatch"
	bypassLAN, _ := json.Marshal(xray.Rules{
		Type:        "field",
		OutboundTag: "direct",
		IP:          []string{"geoip:private"},
	})
	bypassChinaIP, _ := json.Marshal(xray.Rules{
		Type:        "field",
		OutboundTag: "direct",
		IP:          []string{"geoip:cn"},
	})
	bypassChinaSite, _ := json.Marshal(xray.Rules{
		Type:        "field",
		OutboundTag: "direct",
		Domain:      []string{"geosite:cn"},
	})
	blockDomain, _ := json.Marshal(xray.Rules{
		Type:        "field",
		OutboundTag: "blocked",
		Domain:      xray.BlockDomains,
	})
	directDomains, _ := json.Marshal(xray.Rules{
		Type:        "field",
		OutboundTag: "direct",
		Domain:      xray.DirectDomains,
	})
	// blockAd, _ := json.Marshal(v2ray.Rules{
	// 	Type:        "field",
	// 	OutboundTag: "blocked",
	// 	Domain:      []string{"geosite:category-ads-all"},
	// })
	gfwList, _ := json.Marshal(xray.Rules{
		Type:        "field",
		OutboundTag: "proxy",
		Domain:      []string{"geosite:geolocation-!cn"},
	})
	gfwListIP, _ := json.Marshal(xray.Rules{
		Type:        "field",
		OutboundTag: "proxy",
		IP: []string{
			"1.1.1.1/32",
			"1.0.0.1/32",
			"8.8.8.8/32",
			"8.8.4.4/32",
			"149.154.160.0/22",
			"149.154.164.0/22",
			"91.108.4.0/22",
			"91.108.56.0/22",
			"91.108.8.0/22",
			"95.161.64.0/20",
		},
	})
	chinaListSite, _ := json.Marshal(xray.Rules{
		Type:        "field",
		OutboundTag: "proxy",
		Domain:      []string{"geosite:cn"},
	})
	chinaListIP, _ := json.Marshal(xray.Rules{
		Type:        "field",
		OutboundTag: "proxy",
		IP:          []string{"geoip:cn"},
	})
	googleAPI, _ := json.Marshal(xray.Rules{
		Type:        "field",
		OutboundTag: "proxy",
		Domain:      []string{"domain:googleapis.cn", "domain:gstatic.com", "domain:ampproject.org"},
	})
	rules := make([]json.RawMessage, 0)
	if routeMode == 1 {
		rules = []json.RawMessage{
			googleAPI,
			bypassLAN,
			blockDomain,
		}
	}
	if routeMode == 2 {
		rules = []json.RawMessage{
			googleAPI,
			blockDomain,
			bypassChinaSite,
			gfwList,
			bypassChinaIP,
		}
	}
	if routeMode == 3 {
		rules = []json.RawMessage{
			googleAPI,
			blockDomain,
			bypassLAN,
			directDomains,
			bypassChinaSite,
			gfwList, // sniff
			bypassChinaIP,
		}
	}
	if routeMode == 4 {
		rules = []json.RawMessage{
			googleAPI,
			blockDomain,
			gfwListIP,
			gfwList,
		}
	}
	if routeMode == 5 {
		rules = []json.RawMessage{
			// googleAPI,
			blockDomain,
			chinaListSite,
			chinaListIP,
		}
	}
	if routeMode >= 5 {
		rules = []json.RawMessage{
			googleAPI,
			bypassLAN,
			bypassChinaSite,
			bypassChinaIP,
			blockDomain,
			// blockAd,
		}
	}
	return &conf.RouterConfig{
		DomainStrategy: &domainStrategy,
		RuleList:       rules,
	}
}
