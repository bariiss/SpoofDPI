package handler

import (
	"context"
	"net"
	"strconv"

	"github.com/bariiss/SpoofDPI/packet"
	"github.com/bariiss/SpoofDPI/util"
	"github.com/bariiss/SpoofDPI/util/log"
)

type HttpHandler struct {
	bufferSize int
	protocol   string
	port       int
	timeout    int
}

// NewHttpHandler creates a new HttpHandler instance with the given timeout.
func NewHttpHandler(timeout int) *HttpHandler {
	return &HttpHandler{
		bufferSize: 1024,
		protocol:   "HTTP",
		port:       80,
		timeout:    timeout,
	}
}

// Serve handles the HTTP request by establishing a connection to the requested server.
func (h *HttpHandler) Serve(ctx context.Context, lConn *net.TCPConn, pkt *packet.HttpRequest, ip string) {
	ctx = util.GetCtxWithScope(ctx, h.protocol)
	logger := log.GetCtxLogger(ctx)

	// Create a connection to the requested server
	var port = 80
	var err error
	if pkt.Port() != "" {
		port, err = strconv.Atoi(pkt.Port())
		if err != nil {
			logger.Debug().Msgf("error while parsing port for %s aborting..", pkt.Domain())
		}
	}

	rConn, err := net.DialTCP("tcp", nil, &net.TCPAddr{IP: net.ParseIP(ip), Port: port})
	if err != nil {
		err := lConn.Close()
		if err != nil {
			return
		}
		logger.Debug().Msgf("%s", err)
		return
	}

	logger.Debug().Msgf("new connection to the server %s -> %s", rConn.LocalAddr(), pkt.Domain())

	go h.deliverResponse(ctx, rConn, lConn, pkt.Domain(), lConn.RemoteAddr().String())
	go h.deliverRequest(ctx, lConn, rConn, lConn.RemoteAddr().String(), pkt.Domain())

	_, err = rConn.Write(pkt.Raw())
	if err != nil {
		logger.Debug().Msgf("error sending request to %s: %s", pkt.Domain(), err)
		return
	}
}

// deliverRequest reads HTTP requests from the client and forwards them to the server.
func (h *HttpHandler) deliverRequest(ctx context.Context, from *net.TCPConn, to *net.TCPConn, fd string, td string) {
	ctx = util.GetCtxWithScope(ctx, h.protocol)
	logger := log.GetCtxLogger(ctx)

	defer func() {
		err := from.Close()
		if err != nil {
			return
		}
		err = to.Close()
		if err != nil {
			return
		}

		logger.Debug().Msgf("closing proxy connection: %s -> %s", fd, td)
	}()

	for {
		err := setConnectionTimeout(from, h.timeout)
		if err != nil {
			logger.Debug().Msgf("error while setting connection deadline for %s: %s", fd, err)
		}

		pkt, err := packet.ReadHttpRequest(from)
		if err != nil {
			logger.Debug().Msgf("error reading from %s: %s", fd, err)
			return
		}

		pkt.Tidy()

		if _, err := to.Write(pkt.Raw()); err != nil {
			logger.Debug().Msgf("error writing to %s", td)
			return
		}
	}
}

// deliverResponse reads HTTP responses from the server and forwards them to the client.
func (h *HttpHandler) deliverResponse(ctx context.Context, from *net.TCPConn, to *net.TCPConn, fd string, td string) {
	ctx = util.GetCtxWithScope(ctx, h.protocol)
	logger := log.GetCtxLogger(ctx)

	defer func() {
		err := from.Close()
		if err != nil {
			return
		}
		err = to.Close()
		if err != nil {
			return
		}

		logger.Debug().Msgf("closing proxy connection: %s -> %s", fd, td)
	}()

	buf := make([]byte, h.bufferSize)
	for {
		err := setConnectionTimeout(from, h.timeout)
		if err != nil {
			logger.Debug().Msgf("error while setting connection deadline for %s: %s", fd, err)
		}

		bytesRead, err := ReadBytes(from, buf)
		if err != nil {
			logger.Debug().Msgf("error reading from %s: %s", fd, err)
			return
		}

		if _, err := to.Write(bytesRead); err != nil {
			logger.Debug().Msgf("error writing to %s", td)
			return
		}
	}
}
