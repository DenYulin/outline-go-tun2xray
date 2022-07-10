package main

import (
	"github.com/DenYulin/outline-go-tun2xray/outline/xray"
	"os"
)

const (
	NoError                    = 0
	Unexpected                 = 1
	JsonXrayConfigFileNotExist = 2
	StartUpXrayClientFailure   = 3
	OpenTunFailure             = 4
	CopyDataToTunDeviceFailure = 5
)

func fileExists(file string) bool {
	info, err := os.Stat(file)
	return err == nil && !info.IsDir()
}

func toXrayProfile() *xray.Profile {
	return &xray.Profile{
		Host:             *args.host,
		Path:             *args.path,
		TLS:              *args.security,
		Address:          *args.serverAddress,
		Port:             *args.serverPort,
		Net:              *args.net,
		ID:               *args.id,
		Flow:             *args.flow,
		Type:             *args.headerType,
		OutboundProtocol: *args.outboundProtocol,
		UseIPv6:          *args.useIPv6,
		LogLevel:         *args.logLevel,
		RouteMode:        *args.routeMode,
		DNS:              *args.dns,
		AllowInsecure:    *args.allowInsecure,
		Mux:              *args.mux,
		AssetPath:        *args.assetPath,
	}
}
