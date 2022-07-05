package xray

import (
	"bytes"
	"github.com/Jigsaw-Code/outline-go-tun2socks/features"
	"github.com/eycorsican/go-tun2socks/common/log"
	"github.com/xtls/xray-core/core"
	"github.com/xtls/xray-core/infra/conf/serial"
)

type Vless features.VLess

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
