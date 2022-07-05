package common

import (
	"encoding/json"
	"github.com/Jigsaw-Code/outline-go-tun2socks/features"
	"github.com/Jigsaw-Code/outline-go-tun2socks/xray"
	"github.com/eycorsican/go-tun2socks/common/log"
	_ "github.com/xtls/xray-core/common"
	xnet "github.com/xtls/xray-core/common/net"
	"github.com/xtls/xray-core/infra/conf"
	"github.com/xxf098/go-tun2socks-build/v2ray"
	"net"
	"strconv"
	"strings"
)

const SeparatorComma = ","

const (
	VMess       string = "vmess"
	VLess       string = "vless"
	Trojan      string = "trojan"
	ShadowSocks string = "shadowsocks"
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

func CreateVLessOutboundDetourConfig(profile *features.VLess) conf.OutboundDetourConfig {
	// TODO
	return conf.OutboundDetourConfig{}
}

// CreateRouterConfig
// 0 all
// 1 bypass LAN
// 2 bypass China
// 3 bypass LAN & China
// 4 GFWList
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
		Domain:      v2ray.BlockDomains,
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
