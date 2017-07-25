// Package tests includes the various test types.
package tests

import (
	"fmt"
	"net"
	"os"
	"sync"
	"time"
)

// MiddleBox accepts connection, then writes as fast as possible,
// for 5 seconds.
// MSS should be 1456
// Should send data for 5 seconds, then
func MiddleBox(lnr net.Listener, wg *sync.WaitGroup) {
	defer wg.Done()
	_, err := lnr.Accept()
	if err != nil {
		fmt.Println("Error accepting: ", err.Error())
		// TODO - should this be fatal?
		os.Exit(1)
	}

	// TODO - set up MSS and CWND.
	fmt.Println("Middlebox connected")

	timeout := time.After(5 * time.Second)
FOR:
	for {
		select {
		case <-timeout:
			break FOR
		default:
			// Send some data.

		}
	}

	fmt.Println("Done")
}
