// Part of ndt-server-go <https://github.com/m-lab/ndt-server-go>, which
// is free software under the Apache v2.0 License.

package netx

import (
	"errors"
	"net"
	"testing"
	"time"
)

// mockedConn & mockedAddr: mock net.Conn and net.Addr

// mockedAddr implements the net.Addr interface as of golang v1.9.3. If new
// functions are added to net.Addr in the future and are not mocked, attempting
// to access them will most certainly cause a panic at runtime.
type mockedAddr struct {
	network string
	repr    string
}

func (ma mockedAddr) Network() string {
	return ma.network
}

func (ma mockedAddr) String() string {
	return ma.repr
}

// mockedConn implements the net.Conn interface as of golang v1.9.3. If new
// functions are added to net.Conn in the future and are not mocked, attempting
// to access them will most certainly cause a panic at runtime.
type mockedConn struct {
}

func newMockedConn() net.Conn {
	return mockedConn{}
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
	return mockedAddr{
		network: "tcp",
		repr:    "127.0.0.1:54321",
	}
}

func (mockedConn) RemoteAddr() net.Addr {
	return mockedAddr{
		network: "tcp",
		repr:    "127.0.0.1:8080",
	}
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

// TestMockedConnIsUsable ensures that mockedConn implements net.Conn (and
// net.Addr) given the current state of these interfaces as of go v1.9.3.
func TestMockedConnIsUsable(t *testing.T) {
	mc := newMockedConn() // make sure we use it through the interface

	{
		data := make([]byte, 128)
		count, err := mc.Read(data)
		if count != len(data) || err != nil {
			t.Error("mockedConn cannot Read()")
		}
	}

	{
		data := make([]byte, 128)
		count, err := mc.Write(data)
		if count != len(data) || err != nil {
			t.Error("mockedConn cannot Write()")
		}
	}

	{
		err := mc.Close()
		if err != nil {
			t.Error("mockedConn cannot Close()")
		}
	}

	{
		local := mc.LocalAddr()
		// the reason why I'm calling Network() and String is to proof that
		// they are not going to panic, I don't care about values
		if local == nil || local.Network() == "" || local.String() == "" {
			t.Error("mxockedConn cannot LocalAddr()")
		}
	}

	{
		remote := mc.RemoteAddr()
		if remote == nil || remote.Network() == "" || remote.String() == "" {
			t.Error("mxockedConn cannot RemoteAddr()")
		}
	}

	{
		err := mc.SetDeadline(time.Time{})
		if err != nil {
			t.Error("mockedConn cannot SetDeadline()")
		}
	}

	{
		err := mc.SetReadDeadline(time.Time{})
		if err != nil {
			t.Error("mockedConn cannot SetReadDeadline()")
		}
	}

	{
		err := mc.SetWriteDeadline(time.Time{})
		if err != nil {
			t.Error("mockedConn cannot SetWriteDeadline()")
		}
	}
}

// Test: setTimeout

func TestDeadlineConnSetTimeoutWorks(t *testing.T) {
	dc := NewDeadlineConn(mockedConn{})
	err := dc.SetTimeout(-1)
	if err != ErrInvalidTimeout {
		t.Error("we should not be able to set a negative timeout")
	}
	err = dc.SetTimeout(0)
	if err != ErrInvalidTimeout {
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

var errSetReadDeadline = errors.New("err_set_read_deadline")

func (failSetReadDeadline) SetReadDeadline(t time.Time) error {
	return errSetReadDeadline
}

type failRead struct {
	mockedConn
}

var errRead = errors.New("err_read")

func (failRead) Read(base []byte) (int, error) {
	return 0, errRead
}

func TestDeadlineConnReadWorks(t *testing.T) {
	{
		dc := NewDeadlineConn(failSetReadDeadline{})
		data := make([]byte, 128)
		count, err := dc.Read(data)
		if count != 0 || err != errSetReadDeadline {
			t.Error("unexpected return value")
		}
	}

	{
		dc := NewDeadlineConn(failRead{})
		data := make([]byte, 128)
		count, err := dc.Read(data)
		if count != 0 || err != errRead {
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

var errSetWriteDeadline = errors.New("err_set_write_deadline")

func (failSetWriteDeadline) SetWriteDeadline(t time.Time) error {
	return errSetWriteDeadline
}

type failWrite struct {
	mockedConn
}

var errWrite = errors.New("err_write")

func (failWrite) Write(base []byte) (int, error) {
	return 0, errWrite
}

func TestDeadlineConnWriteWorks(t *testing.T) {
	{
		dc := NewDeadlineConn(failSetWriteDeadline{})
		data := make([]byte, 128)
		count, err := dc.Write(data)
		if count != 0 || err != errSetWriteDeadline {
			t.Error("unexpected return value")
		}
	}

	{
		dc := NewDeadlineConn(failWrite{})
		data := make([]byte, 128)
		count, err := dc.Write(data)
		if count != 0 || err != errWrite {
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
