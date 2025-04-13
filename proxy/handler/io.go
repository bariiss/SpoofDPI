package handler

import (
	"errors"
	"fmt"
	"net"
)

// ReadBytes reads bytes from the TCP connection into the provided destination buffer.
// It returns a slice of the data read and any error encountered.
func ReadBytes(conn *net.TCPConn, dest []byte) ([]byte, error) {
	n, err := conn.Read(dest)
	if err != nil {
		var opError *net.OpError
		if errors.As(err, &opError) && opError.Timeout() {
			return dest[:n], fmt.Errorf("read timeout: %w", err)
		}
		return dest[:n], err
	}
	return dest[:n], nil
}
