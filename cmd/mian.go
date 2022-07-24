package main

import (
	"fmt"
	"github.com/DenYulin/outline-go-tun2xray/features"
	"github.com/DenYulin/outline-go-tun2xray/outline/xray"
	_ "github.com/DenYulin/outline-go-tun2xray/xray/features"
	profile2 "github.com/DenYulin/outline-go-tun2xray/xray/profile"
	"github.com/DenYulin/outline-go-tun2xray/xray/tun2xray"
	"github.com/eycorsican/go-tun2socks/common/log"
	"os"
	"os/signal"
	"runtime"
	"runtime/debug"
	"syscall"
)

var profile = &tun2xray.VLess{
	Host:          "www.subaru-rabbit.cc",
	Path:          "/",
	TLS:           "xtls",
	ServerAddress: "www.subaru-rabbit.cc",
	ServerPort:    443,
	Net:           "tcp",
	ID:            "2d45a5ae-eb93-4649-8bd5-fe2c282ed31e", // VLess 的用户 ID
	Flow:          "xtls-rprx-direct",                     // 流控模式，用于选择 XTLS 的算法。
	Type:          "none",                                 // 底层传输设置中 header 的伪装类型。默认"none"
	Protocol:      tun2xray.VLESS,
	VLessOptions:  vLessOptions,
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
	//assetPath := "/usr/local/share/xray"
	//
	//tun2xray.SetLogLevel(profile.LogLevel)
	//
	//err := tun2xray.StartXRayWithTunFd(1, nil, nil, profile, assetPath)
	//if err != nil {
	//	return
	//}

	//startXray()
	startXrayV2()
}

func startXrayV2() {
	//serverAddress := "20.205.36.99"
	//serverPort := uint32(443)
	//userId := "3b7c7324-fee1-452b-91d9-63bebd3b3c09"

	serverAddress := "www.subaru-rabbit.cc"
	serverPort := uint32(443)
	userId := "31e99eb9-1812-4ad1-a442-99c5e9610611"

	allProfile := profile2.CreateProfile(serverAddress, serverPort, userId)
	allProfile.AccessLog = "/tmp/outline/log/access.log"
	allProfile.ErrorLog = "/tmp/outline/log/error.log"
	log.Infof("The param profile: %s", allProfile.String())

	xrayClient, err := tun2xray.StartXRayInstanceWithVLessAndXTLS(allProfile)
	if err != nil {
		log.Errorf("Failed to start xray client with vless and xtls, profile: %s, error: %s", allProfile.String(), err.Error())
		return
	}
	defer xrayClient.Close()

	runtime.GC()
	debug.FreeOSMemory()

	{
		osSignals := make(chan os.Signal, 1)
		signal.Notify(osSignals, os.Interrupt, syscall.SIGTERM)
		<-osSignals
	}
	log.Infof("end")
}

func startXray() {
	logLevel := "debug"
	tun2xray.SetLogLevel(logLevel)

	serverAddress := "20.205.36.99"
	serverPort := 443
	userId := "3b7c7324-fee1-452b-91d9-63bebd3b3c09"

	profile := &xray.Profile{
		InboundPort:      1080,
		Host:             "127.0.0.1",
		ServerAddress:    serverAddress,
		ServerPort:       uint32(serverPort),
		ID:               userId,
		OutboundProtocol: tun2xray.VLESS,
		LogLevel:         logLevel,
		Flow:             "xtls-rprx-direct",
		DNS:              "1.1.1.1:53,8.8.8.8:53,8.8.4.4:53,9.9.9.9:53,208.67.222.222:53",
		Mux:              -1,
		Path:             "/",
		Net:              "tcp",
		TLS:              "xtls",
		Type:             "none",
		AllowInsecure:    false,
		UseIPv6:          false,
	}

	_, xrayErr := xray.CreateXrayClient(profile)
	if xrayErr != nil {
		fmt.Printf("invalid xray proxy parameters, error: %s", xrayErr.Error())
		return
	}

	fmt.Printf("create xray success")
}
