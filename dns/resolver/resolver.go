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
	case dns.TypeA:
		return "A"
	case dns.TypeAAAA:
		return "AAAA"
	default:
		return strconv.FormatUint(uint64(id), 10)
	}
}

// parseAddrsFromMsg extracts IP addresses from a DNS response.
func parseAddrsFromMsg(msg *dns.Msg) []net.IPAddr {
	addrs := make([]net.IPAddr, 0, len(msg.Answer))
	for _, ans := range msg.Answer {
		switch rr := ans.(type) {
		case *dns.A:
			addrs = append(addrs, net.IPAddr{IP: rr.A})
		case *dns.AAAA:
			addrs = append(addrs, net.IPAddr{IP: rr.AAAA})
		}
	}
	return addrs
}

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
func lookupType(ctx context.Context, host string, qType uint16, exchange exchangeFunc) *DNSResult {
	msg := newMsg(host, qType)
	resp, err := exchange(ctx, msg)
	if err != nil {
		err = fmt.Errorf("resolving %s (%s): %w", host, recordTypeIDToName(qType), err)
		return &DNSResult{err: err}
	}
	return &DNSResult{msg: resp}
}

// newMsg creates a DNS message with a question for the specified host and type.
func newMsg(host string, qType uint16) *dns.Msg {
	msg := new(dns.Msg)
	msg.SetQuestion(dns.Fqdn(host), qType)
	return msg
}

// processResults aggregates and sorts successful DNS responses.
func processResults(ctx context.Context, resCh <-chan *DNSResult) ([]net.IPAddr, error) {
	var (
		addrs []net.IPAddr
		errs  []error
	)

	for res := range resCh {
		if res.err != nil {
			errs = append(errs, res.err)
			continue
		}
		addrs = append(addrs, parseAddrsFromMsg(res.msg)...)
	}

	if ctx.Err() != nil {
		return nil, errors.New("context canceled")
	}

	if len(addrs) == 0 {
		return nil, errors.Join(errs...)
	}

	sortAddrs(addrs)
	return addrs, nil
}
