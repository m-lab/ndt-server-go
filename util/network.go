// Part of ndt-server-go <https://github.com/m-lab/ndt-server-go>, which
// is free software under the Apache v2.0 License.

package util

import (
	"bufio"
	"errors"
	"io"
	"net"
	"time"
)

// DeadlineReader is a io.Reader bound to net.Conn. In this reader, the io.Read
// operation will fail if the operation itself takes more than a specific
// timeout to complete. Use SetTimeout to control such timeout.
type DeadlineReader struct {
	io.Reader
	conn    net.Conn
	timeout time.Duration
}

// DefaultReaderTimeout is the default timeout used by a DeadlineReader.
const DefaultReaderTimeout = 10.0 * time.Second

// NewDeadlineReader creates a DeadlineReader bound to |conn|.
func NewDeadlineReader(conn net.Conn) DeadlineReader {
	return DeadlineReader{
		Reader:  bufio.NewReader(conn),
		conn:    conn,
		timeout: DefaultReaderTimeout,
	}
}

// InvalidTimeoutError is returned when you attempt to set an invalid timeout.
var InvalidTimeoutError = errors.New("Timeout is invalid")

// SetTimeout sets the DeadlineReader timeout. Negative and zero timeouts cause
// an InvalidTimeoutError to be retured by this function.
func (dr DeadlineReader) SetTimeout(timeout time.Duration) error {
	if timeout <= 0 {
		return InvalidTimeoutError
	}
	dr.timeout = timeout
	return nil
}

// Read implements reading with a specific timeout.
func (dr DeadlineReader) Read(data []byte) (int, error) {
	count := 0
	err := dr.conn.SetReadDeadline(time.Now().Add(dr.timeout))
	if err != nil {
		return count, err
	}
	count, err = dr.Reader.Read(data)
	if err != nil {
		return count, err
	}
	return count, dr.conn.SetReadDeadline(time.Time{})
}

// DeadlineWriter is like DeadlineReader but for write operations.
type DeadlineWriter struct {
	io.Writer
	conn    net.Conn
	timeout time.Duration
}

// DefaultWriterTimeout is the default timeout used by a DeadlineWriter.
const DefaultWriterTimeout = 10.0 * time.Second

// NewDeadlineWriter creates a DeadlineWriter bound to |conn|.
func NewDeadlineWriter(conn net.Conn) DeadlineWriter {
	return DeadlineWriter{
		Writer:  bufio.NewWriter(conn),
		conn:    conn,
		timeout: DefaultWriterTimeout,
	}
}

// SetTimeout sets the DeadlineWriter timeout. Negative and zero timeouts cause
// an InvalidTimeoutError to be retured by this function.
func (dw DeadlineWriter) SetTimeout(timeout time.Duration) error {
	if timeout <= 0 {
		return InvalidTimeoutError
	}
	dw.timeout = timeout
	return nil
}

// Read implements writing with a specific timeout.
func (dw DeadlineWriter) Write(data []byte) (int, error) {
	count := 0
	err := dw.conn.SetWriteDeadline(time.Now().Add(dw.timeout))
	if err != nil {
		return count, err
	}
	count, err = dw.Writer.Write(data)
	if err != nil {
		return count, err
	}
	return count, dw.conn.SetWriteDeadline(time.Time{})
}
