package handler

import (
	"context"
	"github.com/rs/zerolog"
	"net"
	"regexp"
	"strconv"

	"github.com/bariiss/SpoofDPI/packet"
	"github.com/bariiss/SpoofDPI/util"
	"github.com/bariiss/SpoofDPI/util/log"
)

type HttpsHandler struct {
	bufferSize      int
	protocol        string
	port            int
	timeout         int
	windowsize      int
	exploit         bool
	allowedPatterns []*regexp.Regexp
}

// NewHttpsHandler creates a new HttpsHandler instance with the given timeout, window size, allowed patterns, and exploit flag.
func NewHttpsHandler(
	timeout int,
	windowSize int,
	allowedPatterns []*regexp.Regexp,
	exploit bool,
) *HttpsHandler {
	return &HttpsHandler{
		bufferSize:      1024,
		protocol:        "HTTPS",
		port:            443,
		timeout:         timeout,
		windowsize:      windowSize,
		allowedPatterns: allowedPatterns,
		exploit:         exploit,
	}
}

// Serve handles the HTTPS request by establishing a connection to the requested server.
func (h *HttpsHandler) Serve(
	ctx context.Context,
	lConn *net.TCPConn,
	initPkt *packet.HttpRequest,
	ip string,
) {
	ctx = util.GetCtxWithScope(ctx, h.protocol)
	logger := log.GetCtxLogger(ctx)

	port := 443
	if pktPort := initPkt.Port(); pktPort != "" {
		if parsedPort, err := strconv.Atoi(pktPort); err == nil {
			port = parsedPort
		} else {
			logger.Debug().Msgf("invalid port for %s, using default 443", initPkt.Domain())
		}
	}

	rConn, err := net.DialTCP("tcp", nil, &net.TCPAddr{IP: net.ParseIP(ip), Port: port})
	if err != nil {
		_ = lConn.Close()
		logger.Debug().Msgf("failed to connect to %s: %s", initPkt.Domain(), err)
		return
	}

	logger.Debug().Msgf("new connection to server %s -> %s", rConn.LocalAddr(), initPkt.Domain())

	// Send "200 Connection Established"
	resp := []byte(initPkt.Version() + " 200 Connection Established\r\n\r\n")
	if _, err := lConn.Write(resp); err != nil {
		_ = rConn.Close()
		logger.Debug().Msgf("failed to send 200 to %s: %s", lConn.RemoteAddr(), err)
		return
	}

	logger.Debug().Msgf("sent connection established to %s", lConn.RemoteAddr())

	// Read ClientHello
	m, err := packet.ReadTLSMessage(lConn)
	if err != nil {
		_ = rConn.Close()
		logger.Debug().Msgf("failed to read TLS message from %s: %s", lConn.RemoteAddr(), err)
		return
	}
	if !m.IsClientHello() {
		_ = rConn.Close()
		logger.Debug().Msgf("non-client hello from %s", lConn.RemoteAddr())
		return
	}

	clientHello := m.Raw
	logger.Debug().Msgf("client sent hello %d bytes", len(clientHello))

	// Start communication pipes
	go h.communicate(ctx, rConn, lConn, initPkt.Domain(), lConn.RemoteAddr().String())
	go h.communicate(ctx, lConn, rConn, lConn.RemoteAddr().String(), initPkt.Domain())

	// Send ClientHello (chunked or plain)
	if h.exploit {
		logger.Debug().Msgf("writing chunked client hello to %s", initPkt.Domain())
		chunks := splitInChunks(ctx, clientHello, h.windowsize)
		if _, err := writeChunks(rConn, chunks); err != nil {
			logger.Debug().Msgf("error writing chunked hello to %s: %s", initPkt.Domain(), err)
			return
		}
		return
	}

	logger.Debug().Msgf("writing plain client hello to %s", initPkt.Domain())
	if _, err := rConn.Write(clientHello); err != nil {
		logger.Debug().Msgf("error writing plain hello to %s: %s", initPkt.Domain(), err)
		return
	}
}

// communicate handles the communication between the client and server.
func (h *HttpsHandler) communicate(
	ctx context.Context,
	from, to *net.TCPConn,
	fd, td string,
) {
	ctx = util.GetCtxWithScope(ctx, h.protocol)
	logger := log.GetCtxLogger(ctx)

	defer h.closeBoth(from, to, fd, td, logger)

	buf := make([]byte, h.bufferSize)
	for {
		if err := setConnectionTimeout(from, h.timeout); err != nil {
			logger.Debug().Msgf("timeout error on %s: %s", fd, err)
		}

		n, err := from.Read(buf)
		if err != nil {
			logger.Debug().Msgf("read error from %s: %s", fd, err)
			return
		}

		if _, err := to.Write(buf[:n]); err != nil {
			logger.Debug().Msgf("write error to %s: %s", td, err)
			return
		}
	}
}

// splitInChunks splits the given byte slice into chunks of the specified size.
func splitInChunks(ctx context.Context, data []byte, size int) [][]byte {
	logger := log.GetCtxLogger(ctx)
	logger.Debug().Msgf("window-size: %d", size)

	if size <= 0 {
		logger.Debug().Msg("using legacy fragmentation")
		if len(data) <= 1 {
			return [][]byte{data}
		}
		return [][]byte{data[:1], data[1:]}
	}

	var chunks [][]byte
	for len(data) > 0 {
		chunkSize := size
		if len(data) < size {
			chunkSize = len(data)
		}
		chunks = append(chunks, data[:chunkSize])
		data = data[chunkSize:]
	}
	return chunks
}

// writeChunks writes the given byte slices to the connection.
func writeChunks(conn *net.TCPConn, c [][]byte) (n int, err error) {
	total := 0
	for i := 0; i < len(c); i++ {
		b, err := conn.Write(c[i])
		if err != nil {
			return 0, nil
		}

		total += b
	}

	return total, nil
}

func (h *HttpsHandler) closeBoth(from, to *net.TCPConn, fd, td string, logger zerolog.Logger) {
	if err := from.Close(); err != nil {
		logger.Debug().Msgf("error closing from (%s): %s", fd, err)
	}
	if err := to.Close(); err != nil {
		logger.Debug().Msgf("error closing to (%s): %s", td, err)
	}
	logger.Debug().Msgf("closed proxy connection: %s -> %s", fd, td)
}
