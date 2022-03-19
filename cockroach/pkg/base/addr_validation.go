// Copyright 2018 The Cockroach Authors.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package base

import (
	"context"
	"fmt"
	"net"
	"strconv"

	"github.com/labulakalia/sqlfmt/cockroach/pkg/util/log"
)

// ValidateAddrs controls the address fields in the Config object
// and "fills in" the blanks:
// - the host part of Addr and HTTPAddr is resolved to an IP address
//   if specified (it stays blank if blank to mean "all addresses").
// - the host part of AdvertiseAddr is filled in if blank, either
//   from Addr if non-empty or os.Hostname(). It is also checked
//   for resolvability.
// - non-numeric port numbers are resolved to numeric.
//
// The addresses fields must be guaranteed by the caller to either be
// completely empty, or have both a host part and a port part
// separated by a colon. In the latter case either can be empty to
// indicate it's left unspecified.
func (cfg *Config) ValidateAddrs(ctx context.Context) error {
	return nil
}

// UpdateAddrs updates the listen and advertise port numbers with
// those found during the call to net.Listen().
//
// After ValidateAddrs() the actual listen addr should be equal to the
// one requested; only the port number can change because of
// auto-allocation. We do check this equality here and report a
// warning if any discrepancy is found.
func UpdateAddrs(ctx context.Context, addr, advAddr *string, ln net.Addr) error {
	desiredHost, _, err := net.SplitHostPort(*addr)
	if err != nil {
		return err
	}

	// Update the listen port number and check the actual listen addr is
	// the one requested.
	lnAddr := ln.String()
	lnHost, lnPort, err := net.SplitHostPort(lnAddr)
	if err != nil {
		return err
	}
	requestedAll := (desiredHost == "" || desiredHost == "0.0.0.0" || desiredHost == "::")
	listenedAll := (lnHost == "" || lnHost == "0.0.0.0" || lnHost == "::")
	if (requestedAll && !listenedAll) || (!requestedAll && desiredHost != lnHost) {
		log.Warningf(ctx, "requested to listen on %q, actually listening on %q", desiredHost, lnHost)
	}
	*addr = net.JoinHostPort(lnHost, lnPort)

	// Update the advertised port number if it wasn't set to start
	// with. We don't touch the advertised host, as this may have
	// nothing to do with the listen address.
	advHost, advPort, err := net.SplitHostPort(*advAddr)
	if err != nil {
		return err
	}
	if advPort == "" || advPort == "0" {
		advPort = lnPort
	}
	*advAddr = net.JoinHostPort(advHost, advPort)
	return nil
}

// validateListenAddr validates and normalizes an address suitable for
// use with net.Listen(). This accepts an empty "host" part as "listen
// on all interfaces" and resolves host names to IP addresses.
func validateListenAddr(ctx context.Context, addr, defaultHost string) (string, string, error) {
	host, port, err := getListenAddr(addr, defaultHost)
	if err != nil {
		return "", "", err
	}
	return resolveAddr(ctx, host, port)
}

func getListenAddr(addr, defaultHost string) (string, string, error) {
	host, port := "", ""
	if addr != "" {
		var err error
		host, port, err = net.SplitHostPort(addr)
		if err != nil {
			return "", "", err
		}
	}
	if host == "" {
		host = defaultHost
	}
	if port == "" {
		port = "0"
	}
	return host, port, nil
}

// resolveAddr resolves non-numeric references to numeric references.
func resolveAddr(ctx context.Context, host, port string) (string, string, error) {
	resolver := net.DefaultResolver

	// Resolve the port number. This may translate service names
	// e.g. "postgresql" to a numeric value.
	portNumber, err := resolver.LookupPort(ctx, "tcp", port)
	if err != nil {
		return "", "", err
	}
	port = strconv.Itoa(portNumber)

	// Resolve the address.
	if host == "" {
		// Keep empty. This means "listen on all addresses".
		return host, port, nil
	}

	addr, err := LookupAddr(ctx, resolver, host)
	return addr, port, err
}

// LookupAddr resolves the given address/host to an IP address. If
// multiple addresses are resolved, it returns the first IPv4 address
// available if there is one, otherwise the first address.
func LookupAddr(ctx context.Context, resolver *net.Resolver, host string) (string, error) {
	// Resolve the IP address or hostname to an IP address.
	addrs, err := resolver.LookupIPAddr(ctx, host)
	if err != nil {
		return "", err
	}
	if len(addrs) == 0 {
		return "", fmt.Errorf("cannot resolve %q to an address", host)
	}

	// TODO(knz): the remainder function can be changed to return all
	// resolved addresses once the server is taught to listen on
	// multiple interfaces. #5816

	// LookupIPAddr() can return a mix of IPv6 and IPv4
	// addresses. Conventionally, the first resolved address is
	// "preferred"; however, for compatibility with previous CockroachDB
	// versions, we still prefer an IPv4 address if there is one.
	for _, addr := range addrs {
		if ip := addr.IP.To4(); ip != nil {
			return ip.String(), nil
		}
	}
	// No IPv4 address, return the first resolved address instead.
	return addrs[0].String(), nil
}
