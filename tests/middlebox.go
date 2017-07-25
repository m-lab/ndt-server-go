// Package tests includes the various test types.
package tests

import (
	"net"
)

func MiddleBox(lnr net.Listener, done chan bool) {

	select {
	case <-done:

	}
}
