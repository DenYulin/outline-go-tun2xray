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
	"errors"
	"github.com/DenYulin/outline-go-tun2xray/outline/xray"
	"github.com/eycorsican/go-tun2socks/common/log"
	"runtime/debug"
	"time"
)

func init() {
	// Apple VPN extensions have a memory limit of 15MB. Conserve memory by increasing garbage
	// collection frequency and returning memory to the OS every minute.
	debug.SetGCPercent(10)
	// TODO: Check if this is still needed in go 1.13, which returns memory to the OS
	// automatically.
	ticker := time.NewTicker(time.Minute * 1)
	go func() {
		for range ticker.C {
			debug.FreeOSMemory()
		}
	}()
}

func ConnectXrayTunnel(tunWriter xray.TunWriter, configType, jsonConfig, serverAddress string, serverPort int, userId string) (xray.OutlineTunnel, error) {
	if tunWriter == nil {
		return nil, errors.New("must provide a TunWriter")
	}

	outlineTunnel, err := xray.CreateOutlineTunnel(tunWriter, configType, jsonConfig, serverAddress, serverPort, userId)
	if err != nil {
		log.Errorf("Failed to create outline tunnel, error: %+v", err)
		return nil, err
	}

	return outlineTunnel, nil
}
