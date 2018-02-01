// Part of ndt-server-go <https://github.com/m-lab/ndt-server-go>, which
// is free software under the Apache v2.0 License.

package util

import (
	"errors"
	"net"
	"time"
)

// DeadlineConn is a net.Conn where I/O operations have deadlines. They will
// fail with an error if it takes more than a specific timeout to complete them.
type DeadlineConn struct {
	net.Conn
	timeout time.Duration
}

// DefaultDeadlineConnTimeout is the default timeout used by DeadlineConn.
const DefaultDeadlineConnTimeout = 10.0 * time.Second

// NewDeadlineConn creates a new DeadlineConn.
func NewDeadlineConn(conn net.Conn) DeadlineConn {
	return DeadlineConn{
		Conn: conn,
		timeout: DefaultDeadlineConnTimeout,
	}
}

// InvalidTimeoutError is returned when you attempt to set an invalid timeout.
var InvalidTimeoutError = errors.New("Timeout is invalid")

// SetTimeout sets the DeadlineConn timeout. Negative and zero timeouts cause
// an InvalidTimeoutError to be retured by this function.
func (dc DeadlineConn) SetTimeout(timeout time.Duration) error {
	if timeout <= 0 {
		return InvalidTimeoutError
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
	count, err = dc.Conn.Read(data)
	if err != nil {
		return count, err
	}
	return count, dc.Conn.SetReadDeadline(time.Time{})
}

// Write implements net.Conn.Write with a specific timeout.
func (dc DeadlineConn) Write(data []byte) (int, error) {
	count := 0
	err := dc.Conn.SetWriteDeadline(time.Now().Add(dc.timeout))
	if err != nil {
		return count, err
	}
	count, err = dc.Conn.Write(data)
	if err != nil {
		return count, err
	}
	return count, dc.Conn.SetWriteDeadline(time.Time{})
}
