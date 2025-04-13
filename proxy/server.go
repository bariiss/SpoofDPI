package proxy

import (
	"context"
	"errors"
	"github.com/rs/zerolog"
	"io"
	"net"
	"time"

	"github.com/bariiss/SpoofDPI/util"
	"github.com/bariiss/SpoofDPI/util/log"
)

const (
	BufferSize = 1024
)

var ErrTimeout = errors.New("connection timed out")

// ReadBytes reads bytes from the TCP connection and returns them.
func ReadBytes(conn *net.TCPConn, dest []byte) ([]byte, error) {
	n, err := conn.Read(dest)
	if err != nil {
		var netErr net.Error
		if errors.As(err, &netErr) && netErr.Timeout() {
			return dest[:n], ErrTimeout
		}
		return dest[:n], err
	}
	return dest[:n], nil
}

// Serve pipes data between two TCP connections in one direction.
func Serve(
	ctx context.Context,
	from *net.TCPConn,
	to *net.TCPConn,
	proto string,
	fd string,
	td string,
	timeout int,
) {
	ctx = util.GetCtxWithScope(ctx, proto)
	logger := log.GetCtxLogger(ctx)

	defer closeConnections(from, to, fd, td, logger)

	buf := make([]byte, BufferSize)

	for {
		if timeout > 0 {
			err := from.SetReadDeadline(time.Now().Add(time.Millisecond * time.Duration(timeout)))
			if err != nil {
				logger.Debug().Msgf("error setting timeout for %s: %s", fd, err)
				return
			}
		}

		bytesRead, err := ReadBytes(from, buf)
		if err != nil {
			if errors.Is(err, io.EOF) {
				logger.Debug().Msgf("finished reading from %s", fd)
				return
			}
			logger.Debug().Msgf("read error from %s: %s", fd, err)
			return
		}

		_, err = to.Write(bytesRead)
		if err != nil {
			logger.Debug().Msgf("write error to %s: %s", td, err)
			return
		}
	}
}

// closeConnections closes both TCP connections and logs the result.
func closeConnections(from, to *net.TCPConn, fd, td string, logger zerolog.Logger) {
	if err := from.Close(); err != nil {
		logger.Debug().Msgf("error closing from (%s): %s", fd, err)
	}
	if err := to.Close(); err != nil {
		logger.Debug().Msgf("error closing to (%s): %s", td, err)
	}
	logger.Debug().Msgf("closed proxy connection: %s -> %s", fd, td)
}
