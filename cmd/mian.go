package main

import (
	"github.com/DenYulin/outline-go-tun2xray/features"
	"github.com/DenYulin/outline-go-tun2xray/xray/tun2xray"
)

var profile = &tun2xray.VLess{
	Host:         "www.subaru-rabbit.cc",
	Path:         "/",
	TLS:          "xtls",
	Address:      "www.subaru-rabbit.cc",
	Port:         443,
	Net:          "tcp",
	ID:           "2d45a5ae-eb93-4649-8bd5-fe2c282ed31e", // VLess 的用户 ID
	Flow:         "xtls-rprx-direct",                     // 流控模式，用于选择 XTLS 的算法。
	Type:         "none",                                 // 底层传输设置中 header 的伪装类型。默认"none"
	Protocol:     tun2xray.VLESS,
	VLessOptions: vLessOptions,
}

var vLessOptions = features.VLessOptions{
	UseIPv6:       false, // 是否使用IPv6
	LogLevel:      "debug",
	RouteMode:     0,
	DNS:           "1.1.1.1,8.8.8.8,8.8.4.4,9.9.9.9,208.67.222.222",
	AllowInsecure: false, // 默认值 false
	Mux:           0,     // 最大并发连接数, 默认值8，值范围为 [1, 1024]
}

func main() {
	assetPath := "/usr/local/share/xray"

	tun2xray.SetLogLevel(profile.LogLevel)

	err := tun2xray.StartXRayWithTunFd(1, nil, nil, profile, assetPath)
	if err != nil {
		return
	}
}
