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

package outline

import (
	"github.com/DenYulin/outline-go-tun2xray/tunnel"
	"github.com/eycorsican/go-tun2socks/core"
)

// Tunnel represents a tunnel from a TUN device to a server.
type Tunnel interface {
	tunnel.Tunnel

	// UpdateUDPSupport determines if UDP is supported following a network connectivity change.
	// Sets the tunnel's UDP connection handler accordingly, falling back to DNS over TCP if UDP is not supported.
	// Returns whether UDP proxying is supported in the new network.
	UpdateUDPSupport() bool
}

type outlineTunnel struct {
	tunnel.Tunnel
	lwipStack    core.LWIPStack
	host         string
	port         int
	password     string
	cipher       string
	isUDPEnabled bool // Whether the tunnel supports proxying UDP.
}

// NewTunnel connects a tunnel to a Shadowsocks proxy server and returns an `outline.Tunnel`.
//
// `host` is the IP or domain of the Shadowsocks proxy.
// `port` is the port of the Shadowsocks proxy.
// `password` is the password of the Shadowsocks proxy.
// `cipher` is the encryption cipher used by the Shadowsocks proxy.
// `isUDPEnabled` indicates if the Shadowsocks proxy and the network support proxying UDP traffic.
// `tunWriter` is used to output packets back to the TUN device.  OutlineTunnel.Disconnect() will close `tunWriter`.
//func NewTunnel(host string, port int, password, cipher string, isUDPEnabled bool, tunWriter io.WriteCloser) (Tunnel, error) {
//if tunWriter == nil {
//	return nil, errors.New("must provide a TUN writer")
//}
//
//core.RegisterOutputFn(func(data []byte) (int, error) {
//	return tunWriter.Write(data)
//})
//lwipStack := core.NewLWIPStack()
//base := tunnel.NewTunnel(tunWriter, lwipStack)
//t := &outlineTunnel{base, lwipStack, host, port, password, cipher, isUDPEnabled}
//t.registerConnectionHandlers()
//return t, nil
//}

func (t *outlineTunnel) UpdateUDPSupport() bool {
	return true
}

// Registers UDP and TCP Shadowsocks connection handlers to the tunnel's host and port.
// Registers a DNS/TCP fallback UDP handler when UDP is disabled.
func (t *outlineTunnel) registerConnectionHandlers() {
}
