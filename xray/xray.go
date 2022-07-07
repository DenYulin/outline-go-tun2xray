package xray

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/Jigsaw-Code/outline-go-tun2socks/features"
	"github.com/eycorsican/go-tun2socks/common/log"
	"github.com/xtls/xray-core/common/cmdarg"
	xnet "github.com/xtls/xray-core/common/net"
	"github.com/xtls/xray-core/core"
	"github.com/xtls/xray-core/infra/conf"
	"github.com/xtls/xray-core/infra/conf/serial"
	"net"
	"strconv"
	"strings"
)

const JsonFormat = "json"

type VLess features.VLess

func StartInstance(config []byte) (*core.Instance, error) {
	jsonConfig, err := serial.DecodeJSONConfig(bytes.NewReader(config))
	if err != nil {
		log.Errorf("Decode json config error, error: %s", err.Error())
		return nil, err
	}
	pbConfig, err := jsonConfig.Build()
	if err != nil {
		log.Errorf("Build protobuf config error, error: %s", err.Error())
		return nil, err
	}
	instance, err := core.New(pbConfig)
	if err != nil {
		log.Errorf("Create xray instance with protobuf config error, error: %s", err.Error())
		return nil, err
	}
	if err = instance.Start(); err != nil {
		log.Errorf("Failed to start xray instance, error: %s", err.Error())
		return nil, err
	}

	return instance, nil
}

func StartInstanceWithJson(configFilePath string) (*core.Instance, error) {

	configFiles := cmdarg.Arg{configFilePath}
	config, err := core.LoadConfig(JsonFormat, configFiles)
	if err != nil {
		log.Errorf("Load xray config error, error: %s", err.Error())
		return nil, errors.New("failed to load config files, configFiles: " + configFiles.String())
	}

	{
		jsonBytes, _ := json.Marshal(config)
		log.Infof("Config Json: %s", string(jsonBytes))
	}

	instance, err := core.New(config)
	if err != nil {
		log.Infof("Failed to create xray client, error: %s", err.Error())
		return nil, err
	}

	if err = instance.Start(); err != nil {
		log.Errorf("Failed to start xray instance, error: %s", err.Error())
		return nil, err
	}

	return instance, nil
}

func CreateDNSConfig(option features.VLessOptions) *conf.DNSConfig {
	routeMode := option.RouteMode
	dnsConf := option.DNS
	dns := strings.Split(dnsConf, ",")
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
		Servers: nameServerConfig,
	}
}

func toNameServerConfig(hostPort string) *conf.NameServerConfig {
	host, port, err := net.SplitHostPort(hostPort)
	if err != nil {
		return nil
	}
	p, err := strconv.Atoi(port)
	if err != nil {
		return nil
	}
	newConfig := &conf.NameServerConfig{
		Address: &conf.Address{Address: xnet.ParseAddress(host)},
		Port:    uint16(p),
	}
	return newConfig
}
