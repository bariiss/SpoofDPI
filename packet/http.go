package packet

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
)

var validMethod = map[string]struct{}{
	"DELETE":      {},
	"GET":         {},
	"HEAD":        {},
	"POST":        {},
	"PUT":         {},
	"CONNECT":     {},
	"OPTIONS":     {},
	"TRACE":       {},
	"COPY":        {},
	"LOCK":        {},
	"MKCOL":       {},
	"MOVE":        {},
	"PROPFIND":    {},
	"PROPPATCH":   {},
	"SEARCH":      {},
	"UNLOCK":      {},
	"BIND":        {},
	"REBIND":      {},
	"UNBIND":      {},
	"ACL":         {},
	"REPORT":      {},
	"MKACTIVITY":  {},
	"CHECKOUT":    {},
	"MERGE":       {},
	"M-SEARCH":    {},
	"NOTIFY":      {},
	"SUBSCRIBE":   {},
	"UNSUBSCRIBE": {},
	"PATCH":       {},
	"PURGE":       {},
	"MKCALENDAR":  {},
	"LINK":        {},
	"UNLINK":      {},
}

type HttpRequest struct {
	raw     []byte
	method  string
	domain  string
	port    string
	path    string
	version string
}

// ReadHttpRequest reads an HTTP request from the provided io.Reader.
func ReadHttpRequest(rdr io.Reader) (*HttpRequest, error) {
	p, err := parse(rdr)
	if err != nil {
		return nil, err
	}

	return p, nil
}

func (p *HttpRequest) Raw() []byte {
	return p.raw
}
func (p *HttpRequest) Method() string {
	return p.method
}

func (p *HttpRequest) Domain() string {
	return p.domain
}

func (p *HttpRequest) Port() string {
	return p.port
}

func (p *HttpRequest) Version() string {
	return p.version
}

// IsValidMethod checks if the HTTP method is valid.
func (p *HttpRequest) IsValidMethod() bool {
	if _, exists := validMethod[p.Method()]; exists {
		return true
	}

	return false
}

// IsConnectMethod checks if the HTTP method is CONNECT.
func (p *HttpRequest) IsConnectMethod() bool {
	return p.Method() == "CONNECT"
}

// Tidy removes unnecessary headers and tidies up the HTTP request.
func (p *HttpRequest) Tidy() {
	s := string(p.raw)

	parts := strings.SplitN(s, "\r\n\r\n", 2)
	if len(parts) < 2 {
		// Invalid HTTP request, nothing to tidy
		return
	}

	headers := strings.Split(parts[0], "\r\n")

	// Reconstruct request line
	headers[0] = fmt.Sprintf("%s %s %s", p.method, p.path, p.version)

	var buf bytes.Buffer
	buf.Grow(len(p.raw))

	crlf := []byte("\r\n")
	for _, line := range headers {
		if strings.HasPrefix(line, "Proxy-Connection:") {
			continue // skip this header
		}
		buf.WriteString(line)
		buf.Write(crlf)
	}
	buf.Write(crlf)
	buf.WriteString(parts[1]) // body

	p.raw = buf.Bytes()
}

// parse reads an HTTP request from the provided io.Reader and returns an HttpRequest struct.
func parse(rdr io.Reader) (*HttpRequest, error) {
	sb := strings.Builder{}
	tee := io.TeeReader(rdr, &sb)
	request, err := http.ReadRequest(bufio.NewReader(tee))
	if err != nil {
		return nil, err
	}

	p := &HttpRequest{}
	p.raw = []byte(sb.String())

	p.domain, p.port, err = net.SplitHostPort(request.Host)
	if err != nil {
		p.domain = request.Host
		p.port = ""
	}

	p.method = request.Method
	p.version = request.Proto
	p.path = request.URL.Path

	if request.URL.RawQuery != "" {
		p.path += "?" + request.URL.RawQuery
	}

	if request.URL.RawFragment != "" {
		p.path += "#" + request.URL.RawFragment
	}
	if p.path == "" {
		p.path = "/"
	}

	err = request.Body.Close()
	if err != nil {
		return nil, err
	}
	return p, nil
}
