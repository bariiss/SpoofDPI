package resolver

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strconv"
	"sync"

	"github.com/bariiss/SpoofDPI/dns/addrselect"
	"github.com/miekg/dns"
)

type exchangeFunc = func(ctx context.Context, msg *dns.Msg) (*dns.Msg, error)

type DNSResult struct {
	msg *dns.Msg
	err error
}

// recordTypeIDToName maps DNS record type IDs to their string representations.
func recordTypeIDToName(id uint16) string {
	switch id {
	case 1:
		return "A"
	case 28:
		return "AAAA"
	}
	return strconv.FormatUint(uint64(id), 10)
}

// parseAddrsFromMsg extracts IP addresses from the DNS message.
func parseAddrsFromMsg(msg *dns.Msg) []net.IPAddr {
	var addrs []net.IPAddr

	for _, record := range msg.Answer {
		switch ipRecord := record.(type) {
		case *dns.A:
			addrs = append(addrs, net.IPAddr{IP: ipRecord.A})
		case *dns.AAAA:
			addrs = append(addrs, net.IPAddr{IP: ipRecord.AAAA})
		}
	}
	return addrs
}

// parseIpAddr parses a string into an IP address.
func sortAddrs(addrs []net.IPAddr) {
	addrselect.SortByRFC6724(addrs)
}

// lookupAllTypes performs DNS lookups for all specified types concurrently.
func lookupAllTypes(ctx context.Context, host string, qTypes []uint16, exchange exchangeFunc) <-chan *DNSResult {
	var wg sync.WaitGroup
	resCh := make(chan *DNSResult)

	for _, qType := range qTypes {
		wg.Add(1)
		go func(qType uint16) {
			defer wg.Done()
			select {
			case <-ctx.Done():
				return
			case resCh <- lookupType(ctx, host, qType, exchange):
			}
		}(qType)
	}

	go func() {
		wg.Wait()
		close(resCh)
	}()

	return resCh
}

// lookupType performs a DNS lookup for a specific type and returns the result.
func lookupType(ctx context.Context, host string, queryType uint16, exchange exchangeFunc) *DNSResult {
	msg := newMsg(host, queryType)
	resp, err := exchange(ctx, msg)
	if err != nil {
		queryName := recordTypeIDToName(queryType)
		err = fmt.Errorf("resolving %s, query type %s: %w", host, queryName, err)
		return &DNSResult{err: err}
	}
	return &DNSResult{msg: resp}
}

// newMsg creates a new DNS message with the specified host and query type.
func newMsg(host string, qType uint16) *dns.Msg {
	msg := new(dns.Msg)
	msg.SetQuestion(dns.Fqdn(host), qType)
	return msg
}

// processResults processes the results from the DNS lookups and returns the IP addresses.
func processResults(ctx context.Context, resCh <-chan *DNSResult) ([]net.IPAddr, error) {
	var errs []error
	var addrs []net.IPAddr

	for result := range resCh {
		if result.err != nil {
			errs = append(errs, result.err)
			continue
		}
		resultAddrs := parseAddrsFromMsg(result.msg)
		addrs = append(addrs, resultAddrs...)
	}
	select {
	case <-ctx.Done():
		return nil, errors.New("canceled")
	default:
		if len(addrs) == 0 {
			return addrs, errors.Join(errs...)
		}
	}

	sortAddrs(addrs)
	return addrs, nil
}
