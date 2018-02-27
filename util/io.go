// Part of Neubot <https://neubot.nexacenter.org/>.
// Neubot is free software. See AUTHORS and LICENSE for more
// information on the copying conditions.

package util

// Make sure I/O operations have a deadline
// See <https://groups.google.com/forum/#!topic/golang-nuts/afgEYsoV8j0>

import (
	"errors"
	"net"
	"time"
)

const IoTimeout = 10.0 * time.Second

func IoAccept(listener net.Listener) (net.Conn, error) {
	tcp_listener, okay := listener.(*net.TCPListener)
	if !okay {
		return nil, errors.New("Not a net.TCPListener")
	}
	err := tcp_listener.SetDeadline(time.Now().Add(IoTimeout))
	if err != nil {
		return nil, err
	}
	conn, err := tcp_listener.Accept()
	if err != nil {
		return nil, err
	}
	err = tcp_listener.SetDeadline(time.Time{})
	return conn, err
}
