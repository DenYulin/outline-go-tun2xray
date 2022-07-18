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

package tun2xray

import (
	"github.com/DenYulin/outline-go-tun2xray/outline/xray"
	"github.com/DenYulin/outline-go-tun2xray/tunnel"
	"runtime/debug"

	"github.com/eycorsican/go-tun2socks/common/log"
)

func init() {
	// Conserve memory by increasing garbage collection frequency.
	debug.SetGCPercent(10)
	log.SetLevel(log.DEBUG)
}

func ConnectXrayTunnel(fd int, configType, jsonConfig, serverAddress string, serverPort int, userId string) (xray.OutlineTunnel, error) {
	tun, err := tunnel.MakeTunFile(fd)
	if err != nil {
		log.Errorf("Failed to make a new tun device, fd: %d, error: %+v", fd, err)
		return nil, err
	}

	outlineTunnel, err := xray.CreateOutlineTunnel(tun, configType, jsonConfig, serverAddress, serverPort, userId)
	if err != nil {
		log.Errorf("Failed to create outline tunnel, error: %+v", err)
		return nil, err
	}

	go tunnel.ProcessInputPackets(outlineTunnel, tun)

	return outlineTunnel, nil
}
