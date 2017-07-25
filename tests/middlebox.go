// Package tests includes the various test types.
package tests

import (
	"crypto/rand"
	"fmt"
	"io"
	"net"
	"os"
	"sync"
	"time"

	"github.com/m-lab/ndt-server-go/protocol"
)

// MiddleBox accepts connection, then writes as fast as possible,
// for 5 seconds.
// MSS should be 1456
// Should send data for 5 seconds, then
func MiddleBox(lnr net.Listener, wg *sync.WaitGroup) {
	defer wg.Done()
	data := make([]byte, 8192)
	rand.Read(data)

	mb, err := lnr.Accept()
	if err != nil {
		fmt.Println("Error accepting: ", err.Error())
		// TODO - should this be fatal?
		os.Exit(1)
	}

	// TODO - set up MSS and CWND.
	fmt.Println("Middlebox connected")

	timeout := time.After(5 * time.Second)
	count := 0
FOR:
	for {
		select {
		case <-timeout:
			break FOR
		default:
			// Send some data.
			mb.Write(data)
			count++
		}
	}
	fmt.Println("Total of ", count, " 8KB blocks sent.")
}

func DoMiddleBox(conn io.Writer) {
	// Handle the MiddleBox test.
	lnr, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		fmt.Println("Error listening:", err.Error())
		os.Exit(1)
	}

	_, port, err := net.SplitHostPort(lnr.Addr().String())
	var wg sync.WaitGroup
	wg.Add(1)
	// TODO - is there any real advantage to goroutine here?
	go MiddleBox(lnr, &wg)
	// Send the TEST_PREPARE message.
	protocol.SendJSON(conn, 3, protocol.SimpleMsg{port})
	wg.Wait()
	fmt.Println("Middlebox done")
	protocol.SendJSON(conn, 5, protocol.SimpleMsg{"Results"})
}
