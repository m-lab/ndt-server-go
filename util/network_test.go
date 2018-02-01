// Part of ndt-server-go <https://github.com/m-lab/ndt-server-go>, which
// is free software under the Apache v2.0 License.

package util

import (
	"errors"
	"net"
	"testing"
	"time"
)

// mockedConn: mocks net.Conn

type mockedConn struct {
	net.Conn
}

func (mockedConn) Read(b []byte) (int, error) {
	return len(b), nil
}

func (mockedConn) Write(b []byte) (int, error) {
	return len(b), nil
}

func (mockedConn) Close() error {
	return nil
}

func (mockedConn) LocalAddr() net.Addr {
	return nil // note: for now not required
}

func (mockedConn) RemoteAddr() net.Addr {
	return nil // note: for now not required
}

func (c mockedConn) SetDeadline(t time.Time) error {
	err := c.SetReadDeadline(t)
	if err == nil {
		err = c.SetWriteDeadline(t)
	}
	return err
}

func (mockedConn) SetReadDeadline(t time.Time) error {
	return nil
}

func (mockedConn) SetWriteDeadline(t time.Time) error {
	return nil
}

// Test: setTimeout

func TestDeadlineConnSetTimeoutWorks(t *testing.T) {
	dc := NewDeadlineConn(mockedConn{})
	err := dc.SetTimeout(-1)
	if err != InvalidTimeoutError {
		t.Error("we should not be able to set a negative timeout")
	}
	err = dc.SetTimeout(0)
	if err != InvalidTimeoutError {
		t.Error("we should not be able to set a zero timeout")
	}
	err = dc.SetTimeout(1)
	if err != nil {
		t.Error("we should be able to set a positive timeout")
	}
}

// Test: Read

type failSetReadDeadline struct {
	mockedConn
}

var setReadDeadlineError = errors.New("set_read_deadline_error")

func (failSetReadDeadline) SetReadDeadline(t time.Time) error {
	return setReadDeadlineError
}

type failRead struct {
	mockedConn
}

var readError = errors.New("read_error")

func (failRead) Read(base []byte) (int, error) {
	return 0, readError
}

type failClearReadDeadline struct {
	mockedConn
}

var clearReadDeadlineError = errors.New("clear_read_deadline_error")

func (failClearReadDeadline) SetReadDeadline(t time.Time) error {
	zero := time.Time{}
	if t == zero {
		return clearReadDeadlineError
	}
	return nil
}

func TestDeadlineConnReadWorks(t *testing.T) {
	{
		dc := NewDeadlineConn(failSetReadDeadline{})
		data := make([]byte, 128)
		count, err := dc.Read(data)
		if count != 0 || err != setReadDeadlineError {
			t.Error("unexpected return value")
		}
	}

	{
		dc := NewDeadlineConn(failRead{})
		data := make([]byte, 128)
		count, err := dc.Read(data)
		if count != 0 || err != readError {
			t.Error("unexpected return value")
		}
	}

	{
		dc := NewDeadlineConn(failClearReadDeadline{})
		data := make([]byte, 128)
		count, err := dc.Read(data)
		if count != len(data) || err != clearReadDeadlineError {
			t.Error("unexpected return value")
		}
	}

	{
		dc := NewDeadlineConn(mockedConn{})
		data := make([]byte, 128)
		count, err := dc.Read(data)
		if count != len(data) || err != nil {
			t.Error("unexpected return value")
		}
	}
}

// Test: Write

type failSetWriteDeadline struct {
	mockedConn
}

var setWriteDeadlineError = errors.New("set_write_deadline_error")

func (failSetWriteDeadline) SetWriteDeadline(t time.Time) error {
	return setWriteDeadlineError
}

type failWrite struct {
	mockedConn
}

var writeError = errors.New("write_error")

func (failWrite) Write(base []byte) (int, error) {
	return 0, writeError
}

type failClearWriteDeadline struct {
	mockedConn
}

var clearWriteDeadlineError = errors.New("clear_write_deadline_error")

func (failClearWriteDeadline) SetWriteDeadline(t time.Time) error {
	zero := time.Time{}
	if t == zero {
		return clearWriteDeadlineError
	}
	return nil
}

func TestDeadlineConnWriteWorks(t *testing.T) {
	{
		dc := NewDeadlineConn(failSetWriteDeadline{})
		data := make([]byte, 128)
		count, err := dc.Write(data)
		if count != 0 || err != setWriteDeadlineError {
			t.Error("unexpected return value")
		}
	}

	{
		dc := NewDeadlineConn(failWrite{})
		data := make([]byte, 128)
		count, err := dc.Write(data)
		if count != 0 || err != writeError {
			t.Error("unexpected return value")
		}
	}

	{
		dc := NewDeadlineConn(failClearWriteDeadline{})
		data := make([]byte, 128)
		count, err := dc.Write(data)
		if count != len(data) || err != clearWriteDeadlineError {
			t.Error("unexpected return value")
		}
	}

	{
		dc := NewDeadlineConn(mockedConn{})
		data := make([]byte, 128)
		count, err := dc.Write(data)
		if count != len(data) || err != nil {
			t.Error("unexpected return value")
		}
	}
}
