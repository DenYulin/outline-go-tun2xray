package common

import (
	"errors"
	"github.com/DenYulin/outline-go-tun2xray/outline"
	"github.com/DenYulin/outline-go-tun2xray/outline/xray"
	"github.com/eycorsican/go-tun2socks/common/log"
	"io"
	"os"
	"time"
)

const (
	NoError                    = 0
	Unexpected                 = 1
	NoVPNPermissions           = 2
	AuthenticationFailure      = 3
	UDPConnectivity            = 4
	Unreachable                = 5
	VpnStartFailure            = 6
	IllegalConfiguration       = 7
	JsonXrayConfigFileNotExist = 8
	StartUpXrayClientFailure   = 9
	OpenTunFailure             = 10
	CopyDataToTunDeviceFailure = 11
)

const (
	XRayConfigTypeOfParams = "param"
	XRayConfigTypeOfJson   = "json"
)

const ReachabilityTimeout = 10 * time.Second

// OutlineTunnel embeds the tun2socks.Tunnel interface so it gets exported by gobind.
type OutlineTunnel interface {
	outline.Tunnel
}

// TunWriter is an interface that allows for outputting packets to the TUN (VPN).
type TunWriter interface {
	io.WriteCloser
}

type Args struct {
	TunAddr          *string // tun虚拟设备地址
	TunGw            *string // tun虚拟设备网关
	TunMask          *string // tun虚拟设备地址掩码
	TunName          *string // tun虚拟设备名称，默认都是从 tun0 起。
	TunDNS           *string // tun虚拟设备DNS
	ConfigFormat     *string // 配置格式，json: xray json文件，param: 命令行参数模式
	ConfigFilePath   *string // 配置文件绝对路径
	CheckXrayConfig  *bool   // 是否在检查xray配置是否合规
	Host             *string // 底层传输方式配置中的 host，和
	Path             *string // 底层传输方式配置中的 path or key 路径，默认值为 ["/"]。当有多个值时，每次请求随机选择一个值。
	InboundSocksPort *uint64 // 入站 socks 协议的port
	Security         *string // 是否启用传输层加密，none: 不加密，tls: 使用tls加密，xtls: 使用tls加密
	ServerAddress    *string // 服务器地址，出站Outbound的Address，指向服务端，支持域名、IPv4、IPv6。
	ServerPort       *uint64 // 服务端端口，通常与服务端监听的端口相同
	Net              *string // 底层传输方式，HTTP/2、TCP、WebSocket、QUIC、mKCP、ds、gRPC
	Id               *string // VLESS/VMESS 的用户 ID，可以是任意小于 30 字节的字符串, 也可以是一个合法的 UUID.
	Flow             *string // 流控模式，用于选择 XTLS 的算法。
	HeaderType       *string // 数据包头部伪装设置
	OutboundProtocol *string // 出站协议类型
	UseIPv6          *bool   // 在 Freedom 出站协议中，是否使用IPv6，
	LogLevel         *string // 日志级别
	RouteMode        *int    // 路由模式
	Dns              *string // DNS 地址
	AllowInsecure    *bool   // 是否允许不安全连接（仅用于客户端）。默认值为 false。 当值为 true 时，Xray 不会检查远端主机所提供的 TLS 证书的有效性。
	Mux              *int    // 是否开启 Mux 功能。mux <= 0 不开启；mux > 0 开启，并且设置最大并发连接数为 mux
	AssetPath        *string // xray.location.assetPath
	Version          *bool   // 输出版本号
}

// Profile is the basic parameter used by xray startup
type Profile struct {
	Host             string
	Path             string
	InboundSocksPort uint32
	TLS              string
	Address          string
	Port             uint32
	Net              string
	ID               string
	Flow             string
	Type             string // headerType
	OutboundProtocol string `json:"protocol"`
	UseIPv6          bool   `json:"useIPv6"`
	LogLevel         string `json:"logLevel"`
	RouteMode        int    `json:"routeMode"`
	DNS              string `json:"DNS"`
	AllowInsecure    bool   `json:"allowInsecure"`
	Mux              int    `json:"mux"`
	AssetPath        string `json:"assetPath"`
}

func CreateOutlineTunnel(tun TunWriter, configType, jsonConfig, serverAddress string, serverPort int, userId string) (OutlineTunnel, error) {
	var err error
	var outlineTunnel outline.Tunnel

	if configType == XRayConfigTypeOfParams {
		profile := &Profile{
			Address: serverAddress,
			Port:    uint32(serverPort),
			ID:      userId,
		}

		outlineTunnel, err = xray.NewXrayTunnel(profile, tun)
		if err != nil {
			return nil, err
		}
	} else if configType == XRayConfigTypeOfJson {
		if len(jsonConfig) <= 0 {
			log.Errorf("The param jsonConfig can not be empty")
			return nil, errors.New("param jsonConfig can not be empty")
		}

		outlineTunnel, err = xray.NewXrayTunnelWithJson(jsonConfig, tun)
		if err != nil {
			return nil, err
		}
	}

	return outlineTunnel, nil
}

func FileExists(file string) bool {
	info, err := os.Stat(file)
	return err == nil && !info.IsDir()
}

func ToXrayProfile(args Args) *Profile {
	return &Profile{
		Host:             *args.Host,
		Path:             *args.Path,
		InboundSocksPort: uint32(*args.InboundSocksPort),
		TLS:              *args.Security,
		Address:          *args.ServerAddress,
		Port:             uint32(*args.ServerPort),
		Net:              *args.Net,
		ID:               *args.Id,
		Flow:             *args.Flow,
		Type:             *args.HeaderType,
		OutboundProtocol: *args.OutboundProtocol,
		UseIPv6:          *args.UseIPv6,
		LogLevel:         *args.LogLevel,
		RouteMode:        *args.RouteMode,
		DNS:              *args.Dns,
		AllowInsecure:    *args.AllowInsecure,
		Mux:              *args.Mux,
		AssetPath:        *args.AssetPath,
	}
}
