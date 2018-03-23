package tcpinfo

import (
	"errors"
	"net"
	"syscall"
)

var (
	ErrNoTCPInfoSupport = errors.New("OS does not support TCPInfo")
)

// Alternate (better) way to get tcpinfo
func TCPInfo2(conn *net.TCPConn) (*syscall.TCPInfo, error) {
	return nil, ErrNoTCPInfoSupport
}

// SetMSS uses syscall to set the MSS value on a connection.
func SetMSS(tcp *net.TCPListener, mss int) error {
	return nil
}
