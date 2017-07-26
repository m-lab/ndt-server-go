// Package tests includes the various test types.
package tests

import (
	"crypto/rand"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/m-lab/ndt-server-go/protocol"
)

// MiddleBox accepts connection, then writes as fast as possible,
// for 5 seconds.
// MSS should be 1456
// After connection is accepted should:
//   1. send data for 5 seconds
//   2. Send results
//   3. Receive results
//   4. Finalize
//   5. Close port
//   6. Notify waiter.
func MiddleBox(conn net.Conn, lnr *net.TCPListener, testConn chan net.Conn) {
	data := make([]byte, 8192)
	rand.Read(data)

	mb, err := lnr.AcceptTCP()
	if err != nil {
		fmt.Println("Error accepting: ", err.Error())
		// TODO - should this be fatal?
		os.Exit(1)
	}

	// Without this line, the Sleep above causes hangs at end
	// of send loop.
	// mb.SetWriteBuffer(2 * 8192)
	// TODO - set up MSS and CWND.
	fmt.Println("Middlebox connected")

	start := time.Now()
	count := 0
	mb.SetWriteDeadline(time.Now().Add(5000 * time.Millisecond))
	for {
		// Continously send data.
		_, err := mb.Write(data)
		if err != nil {
			fmt.Println(time.Now().Sub(start), " ", err)
			break
		}
		count++
	}
	fmt.Println("Total of ", count, " 8KB blocks sent.")
	// Send test connection back to be closed.
	testConn <- mb
}

func DoMiddleBox(conn net.Conn) {
	// Handle the MiddleBox test.
	addr, _ := net.ResolveTCPAddr("tcp", "localhost:0")

	lnr, err := net.ListenTCP("tcp", addr)
	if err != nil {
		fmt.Println("Error listening:", err.Error())
		os.Exit(1)
	}
	defer lnr.Close()

	_, port, err := net.SplitHostPort(lnr.Addr().String())
	connChan := make(chan net.Conn)
	go MiddleBox(conn, lnr, connChan)
	// Send the TEST_PREPARE message.
	protocol.SendJSON(conn, 3, protocol.SimpleMsg{port})
	fmt.Println("Waiting for test to complete.")
	testConn := <-connChan
	fmt.Println("Middlebox done")
	protocol.SendJSON(conn, 5, protocol.SimpleMsg{"Results"})
	msg, err := protocol.ReadMessage(conn)
	fmt.Println(string(msg.Content))
	protocol.Send(conn, 6, []byte{})
	testConn.Close()
}
