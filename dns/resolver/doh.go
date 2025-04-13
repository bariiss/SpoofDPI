package resolver

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"regexp"
	"time"

	"github.com/miekg/dns"
)

type DOHResolver struct {
	upstream string
	client   *http.Client
}

// NewDOHResolver creates a new DOHResolver instance.
func NewDOHResolver(host string) *DOHResolver {
	host = regexp.MustCompile(`^https://|/dns-query$`).ReplaceAllString(host, "")
	if ip := net.ParseIP(host); ip != nil && ip.To4() == nil {
		host = fmt.Sprintf("[%s]", ip)
	}

	return &DOHResolver{
		upstream: "https://" + host + "/dns-query",
		client: &http.Client{
			Timeout: 5 * time.Second,
			Transport: &http.Transport{
				DialContext: (&net.Dialer{
					Timeout:   3 * time.Second,
					KeepAlive: 30 * time.Second,
				}).DialContext,
				TLSHandshakeTimeout: 5 * time.Second,
				MaxIdleConnsPerHost: 100,
				MaxIdleConns:        100,
			},
		},
	}
}

// String returns a string representation of the DOHResolver.
func (r *DOHResolver) String() string {
	return fmt.Sprintf("doh resolver(%s)", r.upstream)
}

// Resolve performs a DNS lookup for the given host and returns the IP addresses.
func (r *DOHResolver) Resolve(ctx context.Context, host string, qTypes []uint16) ([]net.IPAddr, error) {
	resultCh := lookupAllTypes(ctx, host, qTypes, r.exchange)
	return processResults(ctx, resultCh)
}

// exchange sends a DNS query to the server and returns the response.
func (r *DOHResolver) exchange(ctx context.Context, msg *dns.Msg) (*dns.Msg, error) {
	packed, err := msg.Pack()
	if err != nil {
		return nil, err
	}

	encoded := base64.RawStdEncoding.EncodeToString(packed)
	url := fmt.Sprintf("%s?dns=%s", r.upstream, encoded)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/dns-message")

	resp, err := r.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			return
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("DoH query failed with status: %s", resp.Status)
	}

	var buf bytes.Buffer
	if _, err := buf.ReadFrom(resp.Body); err != nil {
		return nil, err
	}

	result := new(dns.Msg)
	if err := result.Unpack(buf.Bytes()); err != nil {
		return nil, err
	}

	if result.Rcode != dns.RcodeSuccess {
		return nil, errors.New("doh rcode wasn't successful")
	}

	return result, nil
}
