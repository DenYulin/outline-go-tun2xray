// Copyright 2019 The Outline Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"flag"
	"fmt"
	xrayTunnel "github.com/DenYulin/outline-go-tun2xray/outline/xray"
	"github.com/DenYulin/outline-go-tun2xray/xray"
	"github.com/DenYulin/outline-go-tun2xray/xray/tun2xray"
	"github.com/eycorsican/go-tun2socks/core"
	"github.com/eycorsican/go-tun2socks/tun"
	"github.com/xtls/xray-core/common/session"
	"github.com/xtls/xray-core/common/uuid"
	x2core "github.com/xtls/xray-core/core"
	"io"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/eycorsican/go-tun2socks/common/log"
	_ "github.com/eycorsican/go-tun2socks/common/log/simple" // Register a simple logger.
)

const (
	mtu        = 1500
	udpTimeout = 30 * time.Second
	persistTun = true // Linux: persist the TUN interface after the last open file descriptor is closed.
)

//type Args struct {
//	tunAddr          *string // tun虚拟设备地址
//	tunGw            *string // tun虚拟设备网关
//	tunMask          *string // tun虚拟设备地址掩码
//	tunName          *string // tun虚拟设备名称，默认都是从 tun0 起。
//	tunDNS           *string // tun虚拟设备DNS
//	configFormat     *string // 配置格式，json: xray json文件，param: 命令行参数模式
//	configFilePath   *string // 配置文件绝对路径
//	checkXrayConfig  *bool   // 是否在检查xray配置是否合规
//	host             *string // 底层传输方式配置中的 host，和
//	path             *string // 底层传输方式配置中的 path or key 路径，默认值为 ["/"]。当有多个值时，每次请求随机选择一个值。
//	inboundSocksPort *uint64 // 入站 socks 协议的port
//	security         *string // 是否启用传输层加密，none: 不加密，tls: 使用tls加密，xtls: 使用tls加密
//	serverAddress    *string // 服务器地址，出站Outbound的Address，指向服务端，支持域名、IPv4、IPv6。
//	serverPort       *uint64 // 服务端端口，通常与服务端监听的端口相同
//	net              *string // 底层传输方式，HTTP/2、TCP、WebSocket、QUIC、mKCP、ds、gRPC
//	id               *string // VLESS/VMESS 的用户 ID，可以是任意小于 30 字节的字符串, 也可以是一个合法的 UUID.
//	flow             *string // 流控模式，用于选择 XTLS 的算法。
//	headerType       *string // 数据包头部伪装设置
//	outboundProtocol *string // 出站协议类型
//	useIPv6          *bool   // 在 Freedom 出站协议中，是否使用IPv6，
//	logLevel         *string // 日志级别
//	routeMode        *int    // 路由模式
//	dns              *string // DNS 地址
//	allowInsecure    *bool   // 是否允许不安全连接（仅用于客户端）。默认值为 false。 当值为 true 时，Xray 不会检查远端主机所提供的 TLS 证书的有效性。
//	mux              *int    // 是否开启 Mux 功能。mux <= 0 不开启；mux > 0 开启，并且设置最大并发连接数为 mux
//	assetPath        *string // xray.location.assetPath
//	version          *bool   // 输出版本号
//}

var args = xrayTunnel.Args{}
var version string // Populated at build time through `-X main.version=...`
var lwipWriter io.Writer
var tunDevice io.ReadWriteCloser
var xrayClient *x2core.Instance
var err error

func main() {
	args.TunAddr = flag.String("tunAddr", "10.0.85.2", "TUN interface IP address")
	args.TunGw = flag.String("tunGw", "10.0.85.1", "TUN interface gateway")
	args.TunMask = flag.String("tunMask", "255.255.255.0", "TUN interface network mask; prefixlen for IPv6")
	args.TunDNS = flag.String("tunDNS", "1.1.1.1,9.9.9.9,208.67.222.222", "Comma-separated list of DNS resolvers for the TUN interface (Windows only)")
	args.TunName = flag.String("tunName", "tun0", "TUN interface name")
	args.ConfigFormat = flag.String("configFormat", "json", "")
	args.ConfigFilePath = flag.String("xrayConfigFilePath", "./conf.json", "The xray client conf file path in system")
	args.CheckXrayConfig = flag.Bool("checkXrayConfig", false, "Test xray conf file only, without launching Xray client.")
	args.Host = flag.String("host", "127.0.0.1", "Transport config host")
	args.Path = flag.String("path", "/", "Transport config path")
	args.InboundSocksPort = flag.Uint64("inboundSocksPort", 1080, "inbound socks protocol port")
	args.Security = flag.String("security", "none", "Transport Layer Encryption")
	args.ServerAddress = flag.String("serverAddress", "127.0.0.1", "Server address")
	args.ServerPort = flag.Uint64("serverPort", 443, "Server port")
	args.Net = flag.String("net", "tcp", "transport method")
	args.Id = flag.String("id", getUuid(), "VLess/VMess user id")
	args.Flow = flag.String("flow", "xtls-rprx-direct", "Flow control mode")
	args.HeaderType = flag.String("headerType", "none", "Packet header masquerading settings")
	args.OutboundProtocol = flag.String("outboundProtocol", "vless", "Outbound protocol type")
	args.UseIPv6 = flag.Bool("useIPv6", false, "In freedom protocol, is use ipv6")
	args.LogLevel = flag.String("logLevel", "info", "Logging level: debug|info|warn|error|none")
	args.RouteMode = flag.Int("routeMode", 0, "Route config mode")
	args.Dns = flag.String("dns", "223.5.5.5:53,114.114.114.114:53,8.8.8.8:53,1.1.1.1:53", "Dns address")
	args.AllowInsecure = flag.Bool("allowInsecure", false, "Is allow insecure")
	args.Mux = flag.Int("mux", -1, "Is use mux and maximum number of concurrency")
	args.AssetPath = flag.String("assetPath", "/xray/", "The xray param value of xray.location.assetPath")
	args.Version = flag.Bool("version", false, "Print the version and exit.")
	flag.Parse()

	setLogLevel(*args.LogLevel)

	if *args.Version {
		version = tun2xray.CheckXRayVersion()
		fmt.Println(version)
		os.Exit(0)
	}

	if *args.ConfigFormat == "json" {
		if *args.ConfigFilePath == "" {
			log.Errorf("Must provide a xray config file of json when configFormat is set to json")
			os.Exit(xrayTunnel.JsonXrayConfigFileNotExist)
		} else {
			if !xrayTunnel.FileExists(*args.ConfigFilePath) {
				log.Errorf("The xray config file of json is not exist, configFilePath: %s", *args.ConfigFilePath)
				os.Exit(xrayTunnel.JsonXrayConfigFileNotExist)
			}
		}
		if xrayClient, err = xray.StartInstanceWithJson(*args.ConfigFilePath); err != nil {
			log.Errorf("Failed to start up xray client with json, error: %s", err.Error())
			os.Exit(xrayTunnel.StartUpXrayClientFailure)
		}
	} else if *args.ConfigFormat == "param" {
		profile := xrayTunnel.ToXrayProfile(args)
		xrayClient, err = xrayTunnel.CreateXrayClient(profile)
		if err != nil {
			log.Errorf("Failed to start up xray client with param profile, error: %s", err.Error())
			os.Exit(xrayTunnel.StartUpXrayClientFailure)
		}
	}

	ctx := getContext()

	tunDnsServers := strings.Split(*args.TunDNS, ",")
	if tunDevice, err = tun.OpenTunDevice(*args.TunName, *args.TunAddr, *args.TunGw, *args.TunMask, tunDnsServers, persistTun); err != nil {
		log.Errorf("Failed to open TUN device, error: %s", err.Error())
		os.Exit(xrayTunnel.OpenTunFailure)
	}

	core.RegisterTCPConnHandler(xray.NewTCPHandler(ctx, xrayClient))
	core.RegisterUDPConnHandler(xray.NewUDPHandler(ctx, xrayClient, udpTimeout))
	core.RegisterOutputFn(tunDevice.Write)

	lwipWriter = core.NewLWIPStack()

	go func() {
		if _, err = io.CopyBuffer(lwipWriter, tunDevice, make([]byte, mtu)); err != nil {
			log.Fatalf("Failed to write data to network stack: %v", err)
		}
	}()

	log.Infof("Tun2xray running...")

	osSignals := make(chan os.Signal, 1)
	signal.Notify(osSignals, os.Interrupt, os.Kill, syscall.SIGTERM, syscall.SIGHUP)
	sig := <-osSignals
	log.Debugf("Received signal: %v", sig)
}

func getContext() context.Context {
	ctx := context.Background()
	content := session.ContentFromContext(ctx)
	if content == nil {
		content = new(session.Content)
		ctx = session.ContextWithContent(ctx, content)
	}
	return ctx
}

func setLogLevel(level string) {
	switch strings.ToLower(level) {
	case "debug":
		log.SetLevel(log.DEBUG)
	case "info":
		log.SetLevel(log.INFO)
	case "warn":
		log.SetLevel(log.WARN)
	case "error":
		log.SetLevel(log.ERROR)
	case "none":
		log.SetLevel(log.NONE)
	default:
		log.SetLevel(log.INFO)
	}
}

func getUuid() string {
	u := uuid.New()
	return u.String()
}
