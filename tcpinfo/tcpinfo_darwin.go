package tcpinfo

import (
	"net"
)

// Alternate (better) way to get tcpinfo
func TCPInfo2(conn *net.TCPConn) (string, error) {
	return "No TCPInfo", nil
}

// SetMSS uses syscall to set the MSS value on a connection.
func SetMSS(tcp *net.TCPListener, mss int) {
}
