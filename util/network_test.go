// Part of ndt-server-go <https://github.com/m-lab/ndt-server-go>, which
// is free software under the Apache v2.0 License.

package util

import (
	"testing"
	"errors"
	"io"
	"net"
	"time"
)

// MockedTCPConn implements part of the TCPConn interface for testing purposes
// allowing us to see whether specific methods have been called.
type MockedTCPConn struct {
	ReadCount int
	ReadDeadline time.Time
	ReadError error
	SetReadDeadlineError error
	SetWriteDeadlineError error
	WriteCount int
	WriteDeadline time.Time
	WriteError error
}

func (c MockedTCPConn) SetReadDeadline(t time.Time) error {
	if (c.SetReadDeadlineError != nil) {
		return c.SetReadDeadlineError
	}
	c.ReadDeadline = t
	return nil
}

func (c MockedTCPConn) Read(b []byte) (int, error) {
	return c.ReadCount, c.ReadError
}

func (c MockedTCPConn) SetWriteDeadline(t time.Time) error {
	if (c.SetWriteDeadlineError != nil) {
		return c.SetWriteDeadlineError
	}
	c.WriteDeadline = t
	return nil
}

func (c MockedTCPConn) Write(b []byte) (int, error) {
	return c.WriteCount, c.WriteError
}

// FIXME: timeout zero should remove any timeout for consistency?

// TestDeadlineReaderSetDeadlineWorks ensures that DeadlineReader's SetDeadline
// returns error if passed a negative timeout and nil otherwise.
func TestDeadlineReaderSetDeadlineWorks(t *testing.T) {
	dr := NewDeadlineReader(MockedTCPConn{})
	err := dr.SetDeadline(-1)
	if err != NegativeTimeoutError {
		t.Error("we should not be able to set a negative timeout")
	}
	err = dr.SetDeadline(0)
	if err != nil {
		t.Error("we should not be able to set a zero timeout")
	}
	err = dr.SetDeadline(1)
	if err != nil {
		t.Error("we should not be able to set a positive timeout")
	}
}

// TestDeadlineWriterSetDeadlineWorks ensures that DeadlineWriter's SetDeadline
// returns error if passed a negative timeout and nil otherwise.
func TestDeadlineWriterSetDeadlineWorks(t *testing.T) {
	dr := NewDeadlineWriter(MockedTCPConn{})
	err := dr.SetDeadline(-1)
	if err != NegativeTimeoutError {
		t.Error("we should not be able to set a negative timeout")
	}
	err = dr.SetDeadline(0)
	if err != nil {
		t.Error("we should not be able to set a zero timeout")
	}
	err = dr.SetDeadline(1)
	if err != nil {
		t.Error("we should not be able to set a positive timeout")
	}
}
