// Part of ndt-server-go <https://github.com/m-lab/ndt-server-go>, which
// is free software under the Apache v2.0 License.

// Package netx contains network extensions.
package netx

import (
	"errors"
	"net"
	"time"
)

// DeadlineConn is a net.Conn where I/O operations have deadlines. They will
// fail with an error if it takes more than a specific time to complete them.
type DeadlineConn struct {
	net.Conn
	timeout time.Duration
}

// DefaultTimeout is the default timeout used by DeadlineConn.
const DefaultTimeout = 10.0 * time.Second

// NewDeadlineConn creates a new DeadlineConn.
func NewDeadlineConn(conn net.Conn) DeadlineConn {
	return DeadlineConn{
		Conn:    conn,
		timeout: DefaultTimeout,
	}
}

// ErrInvalidTimeout is returned when you attempt to set an invalid timeout.
var ErrInvalidTimeout = errors.New("Timeout is invalid")

// SetTimeout sets the DeadlineConn timeout. Negative and zero timeouts cause
// an ErrInvalidTimeout to be retured by this function.
func (dc DeadlineConn) SetTimeout(timeout time.Duration) error {
	if timeout <= 0 {
		return ErrInvalidTimeout
	}
	dc.timeout = timeout
	return nil
}

// Read implements net.Conn.Read with a specific timeout.
func (dc DeadlineConn) Read(data []byte) (int, error) {
	count := 0
	err := dc.Conn.SetReadDeadline(time.Now().Add(dc.timeout))
	if err != nil {
		return count, err
	}
	// don't bother with resetting the deadline since it will be set
	// again next time we call Read()
	return dc.Conn.Read(data)
}

// Write implements net.Conn.Write with a specific timeout.
func (dc DeadlineConn) Write(data []byte) (int, error) {
	count := 0
	err := dc.Conn.SetWriteDeadline(time.Now().Add(dc.timeout))
	if err != nil {
		return count, err
	}
	// don't bother with resetting the deadline since it will be set
	// again next time we call Write()
	return dc.Conn.Write(data)
}

// NewTCPListenerWithDeadline constructs a TCPListener that has a specific
// deadline after which all pending Accept()s will fail.
func NewTCPListenerWithDeadline(address string, deadline time.Time) (net.Listener, error) {
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return nil, err
	}
	err = listener.(*net.TCPListener).SetDeadline(deadline)
	if err != nil {
		listener.Close()
		return nil, err
	}
	return listener, nil
}
