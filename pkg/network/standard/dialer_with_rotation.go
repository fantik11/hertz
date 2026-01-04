/*
 * Copyright 2022 CloudWeGo Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package standard

import (
	"crypto/tls"
	"errors"
	"net"
	"sync/atomic"
	"time"

	"github.com/cloudwego/hertz/pkg/network"
)

type RoundRobinRotator struct {
	LocalAddrs    []*net.TCPAddr
	LocalAddrsLen uint32
	rr            uint32
}

func (d *RoundRobinRotator) NewRotator(ips []*net.TCPAddr) {
	d.LocalAddrs = ips
	d.LocalAddrsLen = uint32(len(d.LocalAddrs))
	d.rr = 0
}

func (d *RoundRobinRotator) GetLocalAddr() *net.TCPAddr {
	i := atomic.AddUint32(&d.rr, 1)
	return d.LocalAddrs[i%d.LocalAddrsLen]
}

type dialerWithRotaion struct {
	rotator *RoundRobinRotator
}

func (d *dialerWithRotaion) DialConnection(n, address string, timeout time.Duration, tlsConfig *tls.Config) (conn network.Conn, err error) {
	dn := net.Dialer{Timeout: timeout}
	dn.LocalAddr = d.rotator.GetLocalAddr()
	c, err := dn.Dial(n, address)

	if tlsConfig != nil {
		cTLS := tls.Client(c, tlsConfig)
		conn = newTLSConn(cTLS, defaultMallocSize)
		return
	}
	conn = newConn(c, defaultMallocSize)
	return
}

func (d *dialerWithRotaion) DialTimeout(network, address string, timeout time.Duration, tlsConfig *tls.Config) (conn net.Conn, err error) {
	dn := net.Dialer{Timeout: timeout}
	dn.LocalAddr = d.rotator.GetLocalAddr()

	return dn.Dial(network, address)
}

func (d *dialerWithRotaion) AddTLS(conn network.Conn, tlsConfig *tls.Config) (network.Conn, error) {
	cTlS := tls.Client(conn, tlsConfig)
	err := cTlS.Handshake()
	if err != nil {
		return nil, err
	}
	conn = newTLSConn(cTlS, defaultMallocSize)
	return conn, nil
}

func NewDialerWithRotator(ipList []string) (network.Dialer, error) {
	addrs := make([]*net.TCPAddr, 0, len(ipList))

	for _, s := range ipList {
		ip := net.ParseIP(s)
		if ip == nil {
			continue
		}
		addrs = append(addrs, &net.TCPAddr{IP: ip, Port: 0})
	}

	if len(addrs) == 0 {
		return nil, errors.New("no valid IPs provided")
	}

	rr := &RoundRobinRotator{}
	rr.NewRotator(addrs)

	return &dialerWithRotaion{rotator: rr}, nil
}
