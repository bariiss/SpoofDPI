package handler

import (
	"net"
	"time"
)

// setConnectionTimeout sets the read deadline for the given TCP connection.
func setConnectionTimeout(conn *net.TCPConn, timeout int) error {
	if timeout > 0 {
		return conn.SetReadDeadline(time.Now().Add(time.Millisecond * time.Duration(timeout)))
	}
	return nil
}
