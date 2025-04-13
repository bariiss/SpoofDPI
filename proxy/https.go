package proxy

import (
	"context"
	"net"
	"strconv"

	"github.com/bariiss/SpoofDPI/packet"
	"github.com/bariiss/SpoofDPI/util"
	"github.com/bariiss/SpoofDPI/util/log"
)

const protoHTTPS = "HTTPS"

// handleHttps handles HTTPS requests by establishing a connection to the requested server.
func (pxy *Proxy) handleHttps(
	ctx context.Context,
	lConn *net.TCPConn,
	exploit bool,
	initPkt *packet.HttpRequest,
	ip string,
) {
	ctx = util.GetCtxWithScope(ctx, protoHTTPS)
	logger := log.GetCtxLogger(ctx)

	port := 443
	pktPort := initPkt.Port()
	if pktPort != "" {
		parsedPort, err := strconv.Atoi(pktPort)
		if err != nil {
			logger.Debug().Msgf("invalid port for %s, using default 443", initPkt.Domain())
		}
		if err == nil {
			port = parsedPort
		}
	}

	rConn, err := net.DialTCP("tcp", nil, &net.TCPAddr{IP: net.ParseIP(ip), Port: port})
	if err != nil {
		_ = lConn.Close()
		logger.Debug().Msgf("failed to connect to %s: %s", initPkt.Domain(), err)
		return
	}

	logger.Debug().Msgf("new connection to the server %s -> %s", rConn.LocalAddr(), initPkt.Domain())

	_, err = lConn.Write([]byte(initPkt.Version() + " 200 Connection Established\r\n\r\n"))
	if err != nil {
		_ = rConn.Close()
		_ = lConn.Close()
		logger.Debug().Msgf("failed to send 200 response to client %s: %s", lConn.RemoteAddr(), err)
		return
	}

	logger.Debug().Msgf("sent connection established to %s", lConn.RemoteAddr())

	m, err := packet.ReadTLSMessage(lConn)
	if err != nil {
		_ = rConn.Close()
		_ = lConn.Close()
		logger.Debug().Msgf("error reading TLS message from %s: %s", lConn.RemoteAddr(), err)
		return
	}

	if !m.IsClientHello() {
		_ = rConn.Close()
		_ = lConn.Close()
		logger.Debug().Msgf("received non-client hello from %s", lConn.RemoteAddr())
		return
	}

	clientHello := m.Raw
	logger.Debug().Msgf("client sent hello %d bytes", len(clientHello))

	go Serve(ctx, rConn, lConn, protoHTTPS, initPkt.Domain(), lConn.RemoteAddr().String(), pxy.timeout)

	if exploit {
		logger.Debug().Msgf("writing chunked client hello to %s", initPkt.Domain())
		chunks := splitInChunks(ctx, clientHello, pxy.windowSize)
		_, err := writeChunks(rConn, chunks)
		if err != nil {
			logger.Debug().Msgf("error writing chunked client hello to %s: %s", initPkt.Domain(), err)
			_ = rConn.Close()
			_ = lConn.Close()
			return
		}
	}

	if !exploit {
		logger.Debug().Msgf("writing plain client hello to %s", initPkt.Domain())
		_, err := rConn.Write(clientHello)
		if err != nil {
			logger.Debug().Msgf("error writing plain client hello to %s: %s", initPkt.Domain(), err)
			_ = rConn.Close()
			_ = lConn.Close()
			return
		}
	}

	go Serve(ctx, lConn, rConn, protoHTTPS, lConn.RemoteAddr().String(), initPkt.Domain(), pxy.timeout)
}

// splitInChunks splits the given byte slice into chunks of the specified size.
func splitInChunks(ctx context.Context, bytes []byte, size int) [][]byte {
	logger := log.GetCtxLogger(ctx)

	var chunks [][]byte
	var raw = bytes

	logger.Debug().Msgf("window-size: %d", size)

	if size > 0 {
		for {
			if len(raw) == 0 {
				break
			}

			// necessary check to avoid slicing beyond
			// slice capacity
			if len(raw) < size {
				size = len(raw)
			}

			chunks = append(chunks, raw[0:size])
			raw = raw[size:]
		}

		return chunks
	}

	// When the given window-size <= 0

	if len(raw) < 1 {
		return [][]byte{raw}
	}

	logger.Debug().Msg("using legacy fragmentation")

	return [][]byte{raw[:1], raw[1:]}
}

// writeChunks writes the given byte slices to the connection.
func writeChunks(conn *net.TCPConn, c [][]byte) (int, error) {
	total := 0
	for _, chunk := range c {
		n, err := conn.Write(chunk)
		if err != nil {
			return total, err
		}
		total += n
	}
	return total, nil
}
