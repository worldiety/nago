// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package nethost

import (
	"errors"
	gonet "net"
	"net/http"
	"net/netip"
	"syscall"
	"time"
)

var errForbiddenAddress = errors.New("the address is forbidden")

// metadataAddresses are the addresses of the metadata services of clouds outside of the link-local range.
var metadataAddresses = []netip.Addr{
	netip.MustParseAddr("100.100.100.200"),
	netip.MustParseAddr("168.63.129.16"),
	netip.MustParseAddr("fd00:ec2::254"),
}

var sharedAddressSpace = netip.MustParsePrefix("100.64.0.0/10")

// forbiddenAddress tells, whether a request must not connect to the address.
func forbiddenAddress(ip netip.Addr, allowPrivate bool) bool {
	ip = ip.Unmap()
	switch {
	case !ip.IsValid(), ip.IsUnspecified(), ip.IsMulticast(), ip.IsLinkLocalUnicast(), ip.IsLinkLocalMulticast(), ip.IsInterfaceLocalMulticast():
		return true
	}
	for _, m := range metadataAddresses {
		if ip == m {
			return true
		}
	}
	return !allowPrivate && (ip.IsLoopback() || ip.IsPrivate() || sharedAddressSpace.Contains(ip))
}

// newTransport returns the transport, which each connection decorates. It checks an address after the name was
// resolved, so a name can not point to a forbidden one later, and it uses no proxy of the environment, which would
// connect in its place.
func newTransport(allowPrivate bool) *http.Transport {
	dialer := &gonet.Dialer{
		Timeout: 10 * time.Second,
		Control: func(network, address string, _ syscall.RawConn) error {
			ap, err := netip.ParseAddrPort(address)
			if err != nil || forbiddenAddress(ap.Addr(), allowPrivate) {
				return errForbiddenAddress
			}
			return nil
		},
	}
	return &http.Transport{
		DialContext:         dialer.DialContext,
		ForceAttemptHTTP2:   true,
		MaxIdleConns:        64,
		IdleConnTimeout:     90 * time.Second,
		TLSHandshakeTimeout: 10 * time.Second,
	}
}
