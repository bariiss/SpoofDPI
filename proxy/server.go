package proxy

import (
	"context"
	"errors"
	"io"
	"net"
	"time"

	"github.com/bariiss/SpoofDPI/util"
	"github.com/bariiss/SpoofDPI/util/log"
)

const (
	BufferSize = 1024
)

// ReadBytes reads bytes from the TCP connection and returns them.
func ReadBytes(conn *net.TCPConn, dest []byte) ([]byte, error) {
	n, err := readBytesInternal(conn, dest)
	return dest[:n], err
}

// readBytesInternal reads bytes from the TCP connection into the destination slice.
func readBytesInternal(conn *net.TCPConn, dest []byte) (int, error) {
	totalRead, err := conn.Read(dest)
	if err != nil {
		var opError *net.OpError
		switch {
		case errors.As(err, &opError) && opError.Timeout():
			return totalRead, errors.New("timed out")
		default:
			return totalRead, err
		}
	}
	return totalRead, nil
}

func Serve(ctx context.Context, from *net.TCPConn, to *net.TCPConn, proto string, fd string, td string, timeout int) {
	ctx = util.GetCtxWithScope(ctx, proto)
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

	buf := make([]byte, BufferSize)
	for {
		if timeout > 0 {
			err := from.SetReadDeadline(
				time.Now().Add(time.Millisecond * time.Duration(timeout)),
			)
			if err != nil {
				return
			}
		}

		bytesRead, err := ReadBytes(from, buf)
		if err != nil {
			if err == io.EOF {
				logger.Debug().Msgf("finished reading from %s", fd)
				return
			}
			logger.Debug().Msgf("error reading from %s: %s", fd, err)
			return
		}

		if _, err := to.Write(bytesRead); err != nil {
			logger.Debug().Msgf("error writing to %s", td)
			return
		}
	}
}
