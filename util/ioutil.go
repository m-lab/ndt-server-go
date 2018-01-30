// Part of ndt-server-go <https://github.com/m-lab/ndt-server-go>, which
// is free software under the Apache v2.0 License.

package util

import (
	"io"
	"net"
	"time"
)

// DeadlineReader is a io.Reader where the read has a deadline. See also
// <https://groups.google.com/forum/#!topic/golang-nuts/afgEYsoV8j0>
type DeadlineReader struct {
	io.Reader
	conn     net.Conn
	deadline time.Duration
}

func NewDeadlineReader(conn net.Conn) DeadlineReader {
	return DeadlineReader{
		conn:     conn,
		deadline: 10.0 * time.Second,
	}
}

func (dr DeadlineReader) IoReadFull(data []byte) (int, error) {
	count := 0
	err := dr.conn.SetReadDeadline(time.Now().Add(dr.deadline))
	if err != nil {
		return count, err
	}
	count, err = io.ReadFull(dr.Reader, data)
	if err != nil {
		return count, err
	}
	return count, dr.conn.SetReadDeadline(time.Time{})
}

func (dr DeadlineReader) SetDeadline(deadline time.Duration) {
	dr.deadline = deadline
}
