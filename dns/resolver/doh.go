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
	c := &http.Client{
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
	}

	host = regexp.MustCompile(`^https://|/dns-query$`).ReplaceAllString(host, "")
	if ip := net.ParseIP(host); ip != nil && ip.To4() == nil {
		host = fmt.Sprintf("[%s]", ip)
	}

	return &DOHResolver{
		upstream: "https://" + host + "/dns-query",
		client:   c,
	}
}

// Resolve performs a DNS lookup for the given host and returns the IP addresses.
func (r *DOHResolver) Resolve(ctx context.Context, host string, qTypes []uint16) ([]net.IPAddr, error) {
	resultCh := lookupAllTypes(ctx, host, qTypes, r.exchange)
	addrs, err := processResults(ctx, resultCh)
	return addrs, err
}

// String returns a string representation of the DOHResolver.
func (r *DOHResolver) String() string {
	return fmt.Sprintf("doh resolver(%s)", r.upstream)
}

// exchange sends a DNS query to the server and returns the response.
func (r *DOHResolver) exchange(ctx context.Context, msg *dns.Msg) (*dns.Msg, error) {
	pack, err := msg.Pack()
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s?dns=%s", r.upstream, base64.RawStdEncoding.EncodeToString(pack))
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req = req.WithContext(ctx)
	req.Header.Set("Accept", "application/dns-message")

	resp, err := r.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Println("Error closing response body:", err)
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("doh status error")
	}

	buf := bytes.Buffer{}
	_, err = buf.ReadFrom(resp.Body)
	if err != nil {
		return nil, err
	}

	resultMsg := new(dns.Msg)
	err = resultMsg.Unpack(buf.Bytes())
	if err != nil {
		return nil, err
	}

	if resultMsg.Rcode != dns.RcodeSuccess {
		return nil, errors.New("doh rcode wasn't successful")
	}

	return resultMsg, nil
}
