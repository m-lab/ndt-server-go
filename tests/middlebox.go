// Package tests includes the various test types.
package tests

import (
	"crypto/rand"
	"fmt"
	"net"
	"os"
	"syscall"
	"time"

	"github.com/m-lab/ndt-server-go/protocol"
)

// SetMSS uses syscall to set the MSS value on a connection.
func SetMSS(tcp *net.TCPListener, mss int) {
	file, err := tcp.File()
	if err != nil {
		fmt.Println("error in getting file for the connection!")
		os.Exit(1)
	}
	err = syscall.SetsockoptInt(int(file.Fd()), syscall.SOL_TCP, syscall.TCP_MAXSEG, mss)
	file.Close()
	if err != nil {
		fmt.Println("error in setting MSS option on socket:", err)
		os.Exit(1)
	}
}

// DoMiddleBox listens, accepts connection, then writes as fast as possible,
// for 5 seconds.
// MSS should be 1456
// After connection is accepted should:
//   1. send data for 5 seconds
//   2. Send results
//   3. Receive results
//   4. Finalize
//   5. Close port
//   6. Notify waiter.
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
	// Send the TEST_PREPARE message.
	protocol.SendJSON(conn, 3, protocol.SimpleMsg{Msg: port})
	fmt.Println("Waiting for test to complete.")
	data := make([]byte, 1456)
	rand.Read(data)

	// TODO - set up MSS and CWND.
	SetMSS(lnr, 1456)

	mb, err := lnr.AcceptTCP()
	if err != nil {
		fmt.Println("Error accepting: ", err.Error())
		// TODO - should this be fatal?
		os.Exit(1)
	}

	mb.SetWriteBuffer(8192)
	fmt.Println("Middlebox connected")

	start := time.Now()
	count := 0
	// This deadline actually controls the send period.
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
	fmt.Println("Total of ", count, " ", len(data), " byte blocks sent.")
	// Send test connection back to be closed.
	fmt.Println("Middlebox done")
	protocol.SendJSON(conn, 5, protocol.SimpleMsg{Msg: "Results"})
	msg, err := protocol.ReadMessage(conn)
	fmt.Println(string(msg.Content))
	protocol.Send(conn, 6, []byte{})
	mb.Close()
}
