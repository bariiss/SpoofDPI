package dns

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/bariiss/SpoofDPI/dns/resolver"
	"github.com/bariiss/SpoofDPI/util"
	"github.com/bariiss/SpoofDPI/util/log"
	"github.com/miekg/dns"
)

const scopeDNS = "DNS"

type Resolver interface {
	Resolve(ctx context.Context, host string, qTypes []uint16) ([]net.IPAddr, error)
	String() string
}

type Dns struct {
	host          string
	port          string
	systemClient  Resolver
	generalClient Resolver
	dohClient     Resolver
	qTypes        []uint16
}

// NewDns creates a new Dns instance with the given configuration.
func NewDns(config *util.Config) *Dns {
	addr := config.DnsAddr
	port := strconv.Itoa(config.DnsPort)
	var qTypes []uint16
	if config.DnsIPv4Only {
		qTypes = []uint16{dns.TypeA}
	} else {
		qTypes = []uint16{dns.TypeAAAA, dns.TypeA}
	}
	return &Dns{
		host:          config.DnsAddr,
		port:          port,
		systemClient:  resolver.NewSystemResolver(),
		generalClient: resolver.NewGeneralResolver(net.JoinHostPort(addr, port)),
		dohClient:     resolver.NewDOHResolver(addr),
		qTypes:        qTypes,
	}
}

// ResolveHost resolves the given host using the appropriate resolver based on the configuration.
func (d *Dns) ResolveHost(ctx context.Context, host string, enableDoh, useSystemDns bool) (string, error) {
	ctx = util.GetCtxWithScope(ctx, scopeDNS)
	logger := log.GetCtxLogger(ctx)

	if ip, err := parseIpAddr(host); err == nil {
		return ip.String(), nil
	}

	clt := d.clientFactory(enableDoh, useSystemDns)
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	logger.Debug().Msgf("resolving %s using %s", host, clt)

	start := time.Now()
	addrs, err := clt.Resolve(ctx, host, d.qTypes)
	if err != nil {
		return "", fmt.Errorf("%s: %w", clt, err)
	}

	if len(addrs) == 0 {
		return "", fmt.Errorf("could not resolve %s using %s", host, clt)
	}

	duration := time.Since(start).Milliseconds()
	logger.Debug().Msgf("resolved %s from %s in %d ms", addrs[0].String(), host, duration)

	return addrs[0].String(), nil
}

// clientFactory returns the appropriate resolver based on the configuration.
func (d *Dns) clientFactory(enableDoh, useSystemDns bool) Resolver {
	if useSystemDns {
		return d.systemClient
	}
	if enableDoh {
		return d.dohClient
	}
	return d.generalClient
}

// parseIpAddr parses the given address string into a net.IPAddr.
func parseIpAddr(addr string) (*net.IPAddr, error) {
	ip := net.ParseIP(addr)
	if ip == nil {
		return nil, fmt.Errorf("%s is not an ip address", addr)
	}
	return &net.IPAddr{IP: ip}, nil
}
