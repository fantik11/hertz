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
	"net"
	"sync/atomic"
	"time"

	"github.com/cloudwego/hertz/pkg/network"
)

type dialer struct{}

var (
	srcAddrs = []*net.TCPAddr{
		{IP: net.ParseIP("194.26.229.214")}, // твои исходящие IP
		{IP: net.ParseIP("194.26.229.215")},
		{IP: net.ParseIP("194.26.229.216")},
		{IP: net.ParseIP("194.26.229.217")},
	}
	rr uint64
)

func (d *dialer) DialConnection(n, address string, timeout time.Duration, tlsConfig *tls.Config) (conn network.Conn, err error) {
	//  net.DialTimeout(n, address, timeout)
	dn := net.Dialer{Timeout: timeout}
	c, err := dn.Dial(n, address)

	i := atomic.AddUint64(&rr, 1)
	dn.LocalAddr = srcAddrs[i%uint64(len(srcAddrs))]

	if tlsConfig != nil {
		cTLS := tls.Client(c, tlsConfig)
		conn = newTLSConn(cTLS, defaultMallocSize)
		return
	}
	conn = newConn(c, defaultMallocSize)
	return
}

func (d *dialer) DialTimeout(network, address string, timeout time.Duration, tlsConfig *tls.Config) (conn net.Conn, err error) {
	conn, err = net.DialTimeout(network, address, timeout)
	return
}

func (d *dialer) AddTLS(conn network.Conn, tlsConfig *tls.Config) (network.Conn, error) {
	cTlS := tls.Client(conn, tlsConfig)
	err := cTlS.Handshake()
	if err != nil {
		return nil, err
	}
	conn = newTLSConn(cTlS, defaultMallocSize)
	return conn, nil
}

func NewDialer() network.Dialer {
	return &dialer{}
}
