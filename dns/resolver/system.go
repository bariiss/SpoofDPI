package resolver

import (
	"context"
	"net"
)

type SystemResolver struct {
	*net.Resolver
}

// NewSystemResolver creates a new SystemResolver instance.
func NewSystemResolver() *SystemResolver {
	return &SystemResolver{
		&net.Resolver{PreferGo: true},
	}
}

// LookupIPAddr performs a DNS lookup for the given host and returns the IP addresses.
func (r *SystemResolver) String() string {
	return "system resolver"
}

// LookupIPAddr performs a DNS lookup for the given host and returns the IP addresses.
func (r *SystemResolver) Resolve(ctx context.Context, host string, _ []uint16) ([]net.IPAddr, error) {
	addrs, err := r.LookupIPAddr(ctx, host)
	if err != nil {
		return []net.IPAddr{}, err
	}
	return addrs, nil
}
