package resolver

import (
	"context"
	"net"
)

type SystemResolver struct {
	*net.Resolver
}

// NewSystemResolver creates a new SystemResolver instance using Go's built-in resolver.
func NewSystemResolver() *SystemResolver {
	return &SystemResolver{
		Resolver: &net.Resolver{PreferGo: true},
	}
}

// String returns a string representation of the SystemResolver.
func (r *SystemResolver) String() string {
	return "system resolver"
}

// Resolve performs a DNS lookup for the given host and returns the IP addresses.
func (r *SystemResolver) Resolve(ctx context.Context, host string, _ []uint16) ([]net.IPAddr, error) {
	addrs, err := r.LookupIPAddr(ctx, host)
	if err != nil {
		return nil, err
	}
	return addrs, nil
}
