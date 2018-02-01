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

// DeadlineReader is a net.Conn-based io.Reader where the Read operation has a
// deadline that causes the Reader's io.Read operation to return an error if it
// couldn't return data within said deadline.
type DeadlineReader struct {
	io.Reader
	conn    net.Conn
	timeout time.Duration
}

// DefaultReaderTimeout is the default timeout used by a DeadlineReader.
const DefaultReaderTimeout = 10.0 * time.Second

// NewDeadlineReader creates a DeadlineReader bound to |conn|. The default
// timeout is equal to the value of the DefaultReaderTimeout const.
func NewDeadlineReader(conn net.Conn) DeadlineReader {
	return DeadlineReader{
		Reader:  bufio.NewReader(conn),
		conn:    conn,
		timeout: DefaultReaderTimeout,
	}
}

// NegativeTimeoutError is the error returned when one attempts to set a timeout
// for a deadline reader or writer and such timeout is negative.
const NegativeTimeoutError = errors.New("Timeout is negative")

// SetDeadline sets the  timeout for |dr| to be equal to the specifed |timeout|
// argument. When a io.Read is issued, this |timeout| will be added to the
// current time, to create a deadline for the read operation. This function will
// fail if the specified duration is negative (time.Duration is `int64`).
func (dr DeadlineReader) SetDeadline(timeout time.Duration) error {
	if timeout < 0 {
		return NegativeTimeoutError
	}
	dr.timeout = timeout
	return nil
}

// Read overrides io.Reader's Read method. It sets the read deadline to be equal
// to the currently specified deadline, then issues the real read operation. If
// the Read was successful, the read deadline is then reset using zero since
// "A zero value for t means I/O operations will not time out."
func (dr DeadlineReader) Read(data []byte) (int, error) {
	err := dr.conn.SetReadDeadline(time.Now().Add(dr.timeout))
	if err != nil {
		return count, err
	}
	count, err := Reader.Read(data)
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

// NewDeadlineWriter creates a DeadlineWriter bound to |conn|. The default
// timeout is equal to the value of the DefaultWriterTimeout const.
func NewDeadlineWriter(conn net.Conn) DeadlineWriter {
	return DeadlineWriter{
		Writer:  bufio.NewWriter(conn),
		conn:    conn,
		timeout: DefaultWriterTimeout,
	}
}

// SetDeadline sets the  timeout for |dw| to be equal to the specifed |timeout|
// argument. When a io.Write is issued, this |timeout| will be added to the
// current time, to create a deadline for the write operation. This function
// will fail if the specified duration is negative (time.Duration is `int64`).
func (dw DeadlineWriter) SetDeadline(timeout time.Duration) error {
	if timeout < 0 {
		return NegativeTimeoutError
	}
	dw.timeout = timeout
	return nil
}

// Writer overrides io.Writer's Writer method. It sets the write deadline to be
// equal to the currently specified deadline, then issues the real write
// operation. If the Write was successful, the write deadline is then reset
// using zero since "A zero value for t means I/O operations will not time out."
func (dw DeadlineWriter) Write(data []byte) (int, error) {
	err := dw.conn.SetWriteDeadline(time.Now().Add(dw.timeout))
	if err != nil {
		return count, err
	}
	count, err := Writer.Write(data)
	if err != nil {
		return count, err
	}
	return count, dw.conn.SetWriteDeadline(time.Time{})
}
