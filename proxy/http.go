package proxy

import (
	"context"
	"github.com/bariiss/SpoofDPI/util"
	"net"
	"strconv"

	"github.com/bariiss/SpoofDPI/util/log"

	"github.com/bariiss/SpoofDPI/packet"
)

const protoHTTP = "HTTP"

// handleHttp handles HTTP requests by establishing a connection to the requested server.
func (pxy *Proxy) handleHttp(ctx context.Context, lConn *net.TCPConn, pkt *packet.HttpRequest, ip string) {
	ctx = util.GetCtxWithScope(ctx, protoHTTP)
	logger := log.GetCtxLogger(ctx)

	pkt.Tidy()

	port := 80
	pktPort := pkt.Port()

	if pktPort != "" {
		parsedPort, err := strconv.Atoi(pktPort)
		if err != nil {
			logger.Debug().Msgf("invalid port for %s, using default 80", pkt.Domain())
		}
		if err == nil {
			port = parsedPort
		}
	}

	rConn, err := net.DialTCP("tcp", nil, &net.TCPAddr{
		IP:   net.ParseIP(ip),
		Port: port,
	})
	if err != nil {
		_ = lConn.Close()
		logger.Debug().Msgf("failed to connect to %s: %s", pkt.Domain(), err)
		return
	}

	logger.Debug().Msgf("new connection to server %s -> %s", rConn.LocalAddr(), pkt.Domain())

	go Serve(ctx, rConn, lConn, protoHTTP, pkt.Domain(), lConn.RemoteAddr().String(), pxy.timeout)

	_, err = rConn.Write(pkt.Raw())
	if err != nil {
		logger.Debug().Msgf("error sending request to %s: %s", pkt.Domain(), err)
		_ = rConn.Close()
		_ = lConn.Close()
		return
	}

	logger.Debug().Msgf("sent request to %s", pkt.Domain())

	go Serve(ctx, lConn, rConn, protoHTTP, lConn.RemoteAddr().String(), pkt.Domain(), pxy.timeout)
}
