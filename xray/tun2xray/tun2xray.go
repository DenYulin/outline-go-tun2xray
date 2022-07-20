package tun2xray

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/DenYulin/outline-go-tun2xray/features"
	"github.com/DenYulin/outline-go-tun2xray/pool"
	"github.com/DenYulin/outline-go-tun2xray/xray"
	"github.com/eycorsican/go-tun2socks/common/log"
	t2core "github.com/eycorsican/go-tun2socks/core"
	"github.com/xtls/xray-core/common/bytespool"
	"github.com/xtls/xray-core/common/session"
	"github.com/xtls/xray-core/core"
	"github.com/xtls/xray-core/features/stats"
	"github.com/xtls/xray-core/infra/conf"
	"github.com/xtls/xray-core/infra/conf/serial"
	"github.com/xtls/xray-core/transport/internet"
	"github.com/xxf098/go-tun2socks-build/runner"
	"io"
	"os"
	"strings"
	"time"
)

var err error
var statsManager stats.Manager
var lwipStack t2core.LWIPStack
var xrayInstance *core.Instance
var isStopped = false
var lwipTUNDataPipeTask *runner.Task
var updateStatusPipeTask *runner.Task
var tunDev *pool.Interface
var lwipWriter io.Writer

type VLess features.VLess

type VpnService interface {
	Protect(fd int) bool
}

type QuerySpeed interface {
	UpdateTraffic(up int64, down int64)
}

type PacketFlow interface {
	// WritePacket should write packets to the TUN fd.
	WritePacket(packet []byte)
}

func StartXRayInstanceWithJson(configJson string) (*core.Instance, error) {
	log.Infof("Start up xray with json: %s", configJson)

	config := []byte(configJson)
	if json.Valid(config) {
		log.Errorf("The param configJson is not a valid json string, configJson: %s", configJson)
		return nil, errors.New("configJson format is invalid")
	}

	decodeJSONConfig, err := serial.DecodeJSONConfig(bytes.NewReader(config))
	if err != nil {
		log.Fatalf("Decode json conf with reader error, error: %s", err.Error())
		return nil, err
	}

	pbConfig, err := decodeJSONConfig.Build()
	if err != nil {
		log.Fatalf("Execute decodeJSONConfig.build error, error: %s", err.Error())
		return nil, err
	}
	instance, err := core.New(pbConfig)
	if err != nil {
		return nil, err
	}
	err = instance.Start()
	if err != nil {
		return nil, err
	}
	statsManager = instance.GetFeature(stats.ManagerType()).(stats.Manager)
	return instance, nil
}

func StartXRayInstanceWithVLess(profile *VLess) (*core.Instance, error) {
	config, err := LoadVLessConfig(profile)
	if err != nil {
		return nil, err
	}
	config.DNSConfig = nil

	jsonConfig, err := json.Marshal(config)
	if err != nil {
		log.Fatalf("Config struct deserialize to json bytes error, error: %+v", err)
		return nil, err
	}

	log.Infof("Config struct deserialize to json successfully, jsonConfig: %s", string(jsonConfig))

	decodeJSONConfig, err := serial.DecodeJSONConfig(bytes.NewReader(jsonConfig))
	if err != nil {
		log.Fatalf("Decode json conf with reader error, error: %s", err.Error())
		return nil, err
	}
	decodeJSONConfig.DNSConfig = xray.CreateDNSConfig(profile.VLessOptions)

	pbConfig, err := decodeJSONConfig.Build()
	if err != nil {
		log.Fatalf("Execute decodeJSONConfig.build error, error: %s", err.Error())
		return nil, err
	}
	instance, err := core.New(pbConfig)
	if err != nil {
		log.Errorf("Failed to create a new xray core instance, error: %+v", err)
		return nil, err
	}
	if err = instance.Start(); err != nil {
		log.Errorf("Failed to start xray instance, err: %+v", err)
		return nil, err
	}
	statsManager = instance.GetFeature(stats.ManagerType()).(stats.Manager)
	return instance, nil
}

func StartXRay(packetFlow PacketFlow, vpnService VpnService, querySpeed QuerySpeed, configBytes []byte, assetPath string) error {
	if packetFlow == nil {
		return errors.New("PacketFlow is null")
	}

	if lwipStack == nil {
		// Set up the lwIP stack.
		lwipStack = t2core.NewLWIPStack()
	}

	// Assets
	if err = os.Setenv("xray.location.asset", assetPath); err != nil {
		log.Errorf("Set system environment variable [xray.location.asset] error，error:%s", err.Error())
		return err
	}

	// Protect file descriptors of net connections in the VPN process to prevent infinite loop.
	protectFd := func(s VpnService, fd int) error {
		if s.Protect(fd) {
			return nil
		} else {
			return errors.New(fmt.Sprintf("failed to protect fd %v", fd))
		}
	}
	netCtrl := func(network, address string, fd uintptr) error {
		return protectFd(vpnService, int(fd))
	}
	if err = internet.RegisterDialerController(netCtrl); err != nil {
		log.Errorf("Register dialer controller error, error: %s", err.Error())
		return err
	}
	if err = internet.RegisterListenerController(netCtrl); err != nil {
		log.Errorf("Register listener controller error, error: %s", err.Error())
		return err
	}

	t2core.SetBufferPool(bytespool.GetPool(t2core.BufSize))

	// Start the V2Ray instance.
	xrayInstance, err = xray.StartInstance(configBytes)
	if err != nil {
		log.Errorf("Start xray instance failed: %v", err)
		return err
	}

	// Configure sniffing settings for traffic coming from tun2socks.
	ctx := context.Background()
	content := session.ContentFromContext(ctx)
	if content == nil {
		content = new(session.Content)
		ctx = session.ContextWithContent(ctx, content)
	}

	t2core.RegisterTCPConnHandler(xray.NewTCPHandler(ctx, xrayInstance))
	t2core.RegisterUDPConnHandler(xray.NewUDPHandler(ctx, xrayInstance, 3*time.Minute))

	// Write IP packets back to TUN.
	t2core.RegisterOutputFn(func(data []byte) (int, error) {
		if !isStopped {
			packetFlow.WritePacket(data)
		}
		return len(data), nil
	})

	statsManager = xrayInstance.GetFeature(stats.ManagerType()).(stats.Manager)
	runner.CheckAndStop(updateStatusPipeTask)
	updateStatusPipeTask = createUpdateStatusPipeTask(querySpeed)
	isStopped = false

	return nil
}

func StopXRay() {
	isStopped = true
	if tunDev != nil {
		tunDev.Stop()
	}
	runner.CheckAndStop(updateStatusPipeTask)
	runner.CheckAndStop(lwipTUNDataPipeTask)

	if lwipStack != nil {
		lwipStack.Close()
		lwipStack = nil
	}
	if statsManager != nil {
		statsManager.Close()
		statsManager = nil
	}

	if xrayInstance != nil {
		xrayInstance.Close()
		xrayInstance = nil
	}
	if xrayInstance != nil {
		xrayInstance.Close()
		xrayInstance = nil
	}
}

func StartXRayWithTunFd(tunFd int, vpnService VpnService, querySpeed QuerySpeed, profile *VLess, assetPath string) error {
	tunDev, err = pool.OpenTunDevice(tunFd)
	if err != nil {
		log.Fatalf("failed to open tun device: %v", err)
		return err
	}

	if lwipStack != nil {
		lwipStack.Close()
	}
	lwipStack = t2core.NewLWIPStack()
	lwipWriter = lwipStack.(io.Writer)

	// Assets
	os.Setenv("xray.location.asset", assetPath)

	// Protect file descriptors of net connections in the VPN process to prevent infinite loop.
	protectFd := func(s VpnService, fd int) error {
		if s.Protect(fd) {
			return nil
		} else {
			return errors.New(fmt.Sprintf("failed to protect fd %v", fd))
		}
	}

	netCtrl := func(network, address string, fd uintptr) error {
		return protectFd(vpnService, int(fd))
	}
	internet.RegisterDialerController(netCtrl)
	internet.RegisterListenerController(netCtrl)
	t2core.SetBufferPool(bytespool.GetPool(t2core.BufSize))

	xrayInstance, err = StartXRayInstanceWithVLess(profile)
	if err != nil {
		log.Fatalf("start xray instance failed: %v", err)
		return err
	}

	ctx := context.Background()
	content := session.ContentFromContext(ctx)
	if content == nil {
		content = new(session.Content)
		ctx = session.ContextWithContent(ctx, content)
	}

	// Register tun2socks connection handlers.
	t2core.RegisterTCPConnHandler(xray.NewTCPHandler(ctx, xrayInstance))
	t2core.RegisterUDPConnHandler(xray.NewUDPHandler(ctx, xrayInstance, 3*time.Minute))

	// Write IP packets back to TUN.
	t2core.RegisterOutputFn(func(data []byte) (int, error) {
		return tunDev.Write(data)
	})
	isStopped = false
	runner.CheckAndStop(lwipTUNDataPipeTask)
	runner.CheckAndStop(updateStatusPipeTask)

	lwipTUNDataPipeTask = runner.Go(func(shouldStop runner.S) error {
		zeroErr := errors.New("nil")
		tunDev.Copy(lwipWriter)
		return zeroErr // any errors ?
	})
	updateStatusPipeTask = createUpdateStatusPipeTask(querySpeed)
	return nil
}

func LoadVLessConfig(profile *VLess) (*conf.Config, error) {
	jsonConfig := &conf.Config{}
	jsonConfig.LogConfig = &conf.LogConfig{
		LogLevel: profile.LogLevel,
	}

	// https://github.com/Loyalsoldier/v2ray-rules-dat
	jsonConfig.DNSConfig = CreateDNSConfig(profile.RouteMode, profile.DNS)

	// update rules
	jsonConfig.RouterConfig = CreateRouterConfig(profile.RouteMode)
	jsonConfig.InboundConfigs = []conf.InboundDetourConfig{
		//CreateDokodemoDoorInboundDetourConfig(profile.ServerPort),
		//CreateSocks5InboundDetourConfig(profile.InboundPort),
	}

	//proxyInboundConfig := GetProxyInboundDetourConfig(profile.ServerPort, SOCKS)
	//jsonConfig.InboundConfigs = []conf.InboundDetourConfig{
	//	proxyInboundConfig,
	//}

	proxyOutboundConfig := profile.GetProxyOutboundDetourConfig()

	freedomOutboundDetourConfig := CreateFreedomOutboundDetourConfig(profile.UseIPv6)

	if profile.RouteMode == 4 {
		jsonConfig.OutboundConfigs = []conf.OutboundDetourConfig{
			freedomOutboundDetourConfig,
			proxyOutboundConfig,
		}
	} else {
		jsonConfig.OutboundConfigs = []conf.OutboundDetourConfig{
			proxyOutboundConfig,
			freedomOutboundDetourConfig,
		}
	}

	// policy
	jsonConfig.Policy = CreatePolicyConfig()
	// stats
	jsonConfig.Stats = &conf.StatsConfig{}
	return jsonConfig, nil
}

func GetProxyInboundDetourConfig(proxyPort uint32, protocol string) conf.InboundDetourConfig {
	proxyInboundConfig := conf.InboundDetourConfig{}

	if protocol == SOCKS {
		proxyInboundConfig = CreateSocks5InboundDetourConfig(proxyPort)
	} else if protocol == DOKODEMO_DOOR {
		proxyInboundConfig = CreateDokodemoDoorInboundDetourConfig(proxyPort)
	}

	return proxyInboundConfig
}

func (profile *VLess) GetProxyOutboundDetourConfig() conf.OutboundDetourConfig {
	proxyOutboundConfig := conf.OutboundDetourConfig{}
	if profile.Protocol == VLESS {
		proxyOutboundConfig.Tag = "vless-out"
		proxyOutboundConfig = CreateVLessOutboundDetourConfig(profile)
	}

	return proxyOutboundConfig
}

func createUpdateStatusPipeTask(querySpeed QuerySpeed) *runner.Task {
	return runner.Go(func(shouldStop runner.S) error {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()
		zeroErr := errors.New("nil")
		for {
			if shouldStop() {
				break
			}
			select {
			case <-ticker.C:
				up := QueryOutboundStats("proxy", "uplink")
				down := QueryOutboundStats("proxy", "downlink")
				querySpeed.UpdateTraffic(up, down)
				// case <-lwipTUNDataPipeTask.StopChan():
				// 	return errors.New("stopped")
			}
		}
		return zeroErr
	})
}

func QueryOutboundStats(tag string, direct string) int64 {
	if statsManager == nil {
		return QueryOutboundXStats(tag, direct)
	}
	counter := statsManager.GetCounter(fmt.Sprintf("outbound>>>%s>>>traffic>>>%s", tag, direct))
	if counter == nil {
		return 0
	}
	return counter.Set(0)
}

func QueryOutboundXStats(tag string, direct string) int64 {
	if statsManager == nil {
		return 0
	}
	counter := statsManager.GetCounter(fmt.Sprintf("outbound>>>%s>>>traffic>>>%s", tag, direct))
	if counter == nil {
		return 0
	}
	return counter.Set(0)
}

func CheckXRayVersion() string {
	return core.Version()
}

func SetLogLevel(logLevel string) {
	// Set log level.
	switch strings.ToLower(logLevel) {
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
		panic("unsupported logging level")
	}
}
