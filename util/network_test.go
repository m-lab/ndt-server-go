// Part of ndt-server-go <https://github.com/m-lab/ndt-server-go>, which
// is free software under the Apache v2.0 License.

package util

import (
	"errors"
	"net"
	"testing"
	"time"
)

// MockedConn mocks net.Conn.
type MockedConn struct {
}

// Read mocks net.Conn.Read.
func (c MockedConn) Read(b []byte) (int, error) {
	return len(b), nil
}

// Write mocks net.Conn.Write.
func (c MockedConn) Write(b []byte) (int, error) {
	return len(b), nil
}

// Close mocks net.Conn.Close.
func (c MockedConn) Close() error {
	return nil
}

// LocalAddr mocks net.Conn.LocalAddr.
func (c MockedConn) LocalAddr() net.Addr {
	return nil // note: for now not required
}

// RemoteAddr mocks net.Conn.RemoteAddr.
func (c MockedConn) RemoteAddr() net.Addr {
	return nil // note: for now not required
}

// SetDeadline mocks net.Conn.SetDeadline.
func (c MockedConn) SetDeadline(t time.Time) error {
	err := c.SetReadDeadline(t)
	if err == nil {
		err = c.SetWriteDeadline(t)
	}
	return err
}

// SetReadDeadline mocks net.Conn.SetReadDeadline.
func (c MockedConn) SetReadDeadline(t time.Time) error {
	return nil
}

// SetWriteDeadline mocks net.Conn.SetWriteDeadline.
func (c MockedConn) SetWriteDeadline(t time.Time) error {
	return nil
}

// TestDeadlineReaderSetTimeoutWorks ensures that DeadlineReader's SetTimeout
// returns error if passed a negative timeout and nil otherwise.
func TestDeadlineReaderSetTimeoutWorks(t *testing.T) {
	dr := NewDeadlineReader(MockedConn{})
	err := dr.SetTimeout(-1)
	if err != InvalidTimeoutError {
		t.Error("we should not be able to set a negative timeout")
	}
	err = dr.SetTimeout(0)
	if err != InvalidTimeoutError {
		t.Error("we should not be able to set a zero timeout")
	}
	err = dr.SetTimeout(1)
	if err != nil {
		t.Error("we should be able to set a positive timeout")
	}
}

// FailSetReadDeadline is a MockedConn where SetReadDeadline fails.
type FailSetReadDeadline struct {
	MockedConn
}

// SetReadDeadlineError is the error returned by a failed SetReadDeadline.
var SetReadDeadlineError = errors.New("mocked error")

// SetReadDeadline always returns an error.
func (c FailSetReadDeadline) SetReadDeadline(t time.Time) error {
	return SetReadDeadlineError
}

// TestDeadlineReaderSetReadDeadlineErrorHandled ensures that we correctly
// deal with a SetReadDeadline error.
func TestDeadlineReaderSetReadDeadlineErrorHandled(t *testing.T) {
	dr := NewDeadlineReader(FailSetReadDeadline{})
	data := make([]byte, 128)
	count, err := dr.Read(data)
	if count != 0 {
		t.Error("unexpected count value returned")
	}
	if err != SetReadDeadlineError {
		t.Error("unexpected error returned")
	}
}

// FailRead is a MockedConn where Read fails.
type FailRead struct {
	MockedConn
}

// ReadError is the error returned by a failed Read.
var ReadError = errors.New("mocked error")

// Read always returns an error.
func (c FailRead) Read(b []byte) (int, error) {
	return 0, ReadError
}

// TestDeadlineReaderReadErrorHandled ensures that we correctly deal with a
// Read error. It also ensures that the deadline has been changed.
func TestDeadlineReaderReadErrorHandled(t *testing.T) {
	dr := NewDeadlineReader(FailRead{})
	data := make([]byte, 128)
	count, err := dr.Read(data)
	if count != 0 {
		t.Error("unexpected count value returned")
	}
	if err != ReadError {
		t.Error("unexpected error returned")
	}
}

// FailResetReadDeadline is a MockedConn where the a SetReadDeadline that
// passes in a zero read deadline will fail.
type FailResetReadDeadline struct {
	MockedConn
}

// FailResetReadDeadlineError is the error returned by a failed second
// SetReadDeadline (meaning that the first succeeds).
var FailResetReadDeadlineError = errors.New("mocked error")

// SetReadDeadline succeeds unless passed a zero deadline.
func (c FailResetReadDeadline) SetReadDeadline(t time.Time) error {
	zero := time.Time{}
	if t == zero {
		return FailResetReadDeadlineError
	}
	return nil
}

// TestDeadlineReaderFailResetReadDeadlineErrorHandled ensures that we correctly
// deal with an error in the second SetReadDeadline.
func TestDeadlineReaderFailResetReadDeadlineErrorHandled(t *testing.T) {
	dr := NewDeadlineReader(FailResetReadDeadline{})
	data := make([]byte, 128)
	count, err := dr.Read(data)
	if count != len(data) {
		t.Error("unexpected count value returned")
	}
	if err != FailResetReadDeadlineError {
		t.Error("unexpected error returned")
	}
}

// TestDeadlineReaderWorks make sure that DeadlineReader works.
func TestDeadlineReaderWorks(t *testing.T) {
	dr := NewDeadlineReader(MockedConn{})
	data := make([]byte, 128)
	count, err := dr.Read(data)
	if count != len(data) {
		t.Error("unexpected count value returned")
	}
	if err != nil {
		t.Error("unexpected error returned")
	}
}

// TestDeadlineWriterSetTimeoutWorks ensures that DeadlineWriter's SetTimeout
// returns error if passed a negative timeout and nil otherwise.
func TestDeadlineWriterSetTimeoutWorks(t *testing.T) {
	dw := NewDeadlineWriter(MockedConn{})
	err := dw.SetTimeout(-1)
	if err != InvalidTimeoutError {
		t.Error("we should not be able to set a negative timeout")
	}
	err = dw.SetTimeout(0)
	if err != InvalidTimeoutError {
		t.Error("we should not be able to set a zero timeout")
	}
	err = dw.SetTimeout(1)
	if err != nil {
		t.Error("we should be able to set a positive timeout")
	}
}

// FailSetWriteDeadline is a MockedConn where SetWriteDeadline fails.
type FailSetWriteDeadline struct {
	MockedConn
}

// SetWriteDeadlineError is the error returned by a failed SetWriteDeadline.
var SetWriteDeadlineError = errors.New("mocked error")

// SetWriteDeadline always returns an error.
func (c FailSetWriteDeadline) SetWriteDeadline(t time.Time) error {
	return SetWriteDeadlineError
}

// TestDeadlineWriterSetWriteDeadlineErrorHandled ensures that we correctly
// deal with a SetWriteDeadline error.
func TestDeadlineWriterSetWriteDeadlineErrorHandled(t *testing.T) {
	dw := NewDeadlineWriter(FailSetWriteDeadline{})
	data := make([]byte, 128)
	count, err := dw.Write(data)
	if count != 0 {
		t.Error("unexpected count value returned")
	}
	if err != SetWriteDeadlineError {
		t.Error("unexpected error returned")
	}
}

// FailWrite is a MockedConn where Write fails.
type FailWrite struct {
	MockedConn
}

// WriteError is the error returned by a failed Write.
var WriteError = errors.New("mocked error")

// Write always returns an error.
func (c FailWrite) Write(b []byte) (int, error) {
	return 0, WriteError
}

// TestDeadlineWriterWriteErrorHandled ensures that we correctly deal with a
// Write error. It also ensures that the deadline has been changed.
func TestDeadlineWriterWriteErrorHandled(t *testing.T) {
	dw := NewDeadlineWriter(FailWrite{})
	data := make([]byte, 128)
	count, err := dw.Write(data)
	t.Log(count)
	if count != 0 {
		t.Error("unexpected count value returned")
	}
	t.Log(err)
	if err != WriteError {
		t.Error("unexpected error returned")
	}
}

// FailResetWriteDeadline is a MockedConn where the a SetWriteDeadline that
// passes in a zero write deadline will fail.
type FailResetWriteDeadline struct {
	MockedConn
}

// FailResetWriteDeadlineError is the error returned by a failed second
// SetWriteDeadline (meaning that the first succeeds).
var FailResetWriteDeadlineError = errors.New("mocked error")

// SetWriteDeadline succeeds unless passed a zero deadline.
func (c FailResetWriteDeadline) SetWriteDeadline(t time.Time) error {
	zero := time.Time{}
	if t == zero {
		return FailResetWriteDeadlineError
	}
	return nil
}

// TestDeadlineWriterFailResetWriteDeadlineErrorHandled ensures that we
// correctly deal with an error in the second SetWriteDeadline.
func TestDeadlineWriterFailResetWriteDeadlineErrorHandled(t *testing.T) {
	dw := NewDeadlineWriter(FailResetWriteDeadline{})
	data := make([]byte, 128)
	count, err := dw.Write(data)
	if count != len(data) {
		t.Error("unexpected count value returned")
	}
	if err != FailResetWriteDeadlineError {
		t.Error("unexpected error returned")
	}
}

// TestDeadlineWriterWorks make sure that DeadlineWriter works.
func TestDeadlineWriterWorks(t *testing.T) {
	dw := NewDeadlineWriter(MockedConn{})
	data := make([]byte, 128)
	count, err := dw.Write(data)
	if count != len(data) {
		t.Error("unexpected count value returned")
	}
	if err != nil {
		t.Error("unexpected error returned")
	}
}
