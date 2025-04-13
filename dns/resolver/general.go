package resolver

import (
	"context"
	"fmt"
	"net"

	"github.com/miekg/dns"
)

type GeneralResolver struct {
	client *dns.Client
	server string
}

// NewGeneralResolver creates a new GeneralResolver instance.
func NewGeneralResolver(server string) *GeneralResolver {
	return &GeneralResolver{
		client: &dns.Client{},
		server: server,
	}
}

// String returns a string representation of the GeneralResolver.
func (r *GeneralResolver) String() string {
	return fmt.Sprintf("general resolver(%s)", r.server)
}

// Resolve performs a DNS lookup for the given host and returns the IP addresses.
func (r *GeneralResolver) Resolve(ctx context.Context, host string, qTypes []uint16) ([]net.IPAddr, error) {
	resultCh := lookupAllTypes(ctx, host, qTypes, r.exchange)
	return processResults(ctx, resultCh)
}

// exchange sends a DNS query to the server and returns the response.
func (r *GeneralResolver) exchange(_ context.Context, msg *dns.Msg) (*dns.Msg, error) {
	resp, _, err := r.client.Exchange(msg, r.server)
	return resp, err
}
