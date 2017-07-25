package main

/*
MSG_LOGIN uses binary protocol
MSG_EXTENDED_LOGIN uses binary message types, but json message bodies.


*/

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"sync"

	"github.com/m-lab/ndt-server-go/protocol"
	"github.com/m-lab/ndt-server-go/tests"
)

const (
	HOST = "localhost"
	PORT = "3001"
	TYPE = "tcp"
)

// returns the port for the middlebox test.
func startMiddleBox() (string, *sync.WaitGroup) {
	// Handle the MiddleBox test.
	mbox, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		fmt.Println("Error listening:", err.Error())
		os.Exit(1)
	}

	_, port, err := net.SplitHostPort(mbox.Addr().String())
	var wg sync.WaitGroup
	wg.Add(1)
	go tests.MiddleBox(mbox, &wg)
	return port, &wg
}

func handleRequest(conn net.Conn) {
	// Close the connection when you're done with it.
	defer conn.Close()
	rdr := bufio.NewReader(conn)

	// Read the incoming login message.
	login, err := protocol.ReadLogin(rdr)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(login)
	// Send "Kickoff" message
	conn.Write([]byte("123456 654321"))
	// Send next messages in the handshake.
	protocol.SendJSON(conn, 1, protocol.SimpleMsg{"0"})
	protocol.SendJSON(conn, 2, protocol.SimpleMsg{"v3.8.1"})

	// TODO - this should be in response to the actual request.
	protocol.SendJSON(conn, 2, protocol.SimpleMsg{"1 2 4 8 32"})

	port, wg := startMiddleBox()

	// Send the TEST_PREPARE message.
	protocol.SendJSON(conn, 3, protocol.SimpleMsg{port})
	wg.Wait()
	fmt.Println("Middlebox done")

	protocol.ReadMessage(rdr)
}

func main() {
	l, err := net.Listen(TYPE, HOST+":"+PORT)
	if err != nil {
		fmt.Println("Error listening:", err.Error())
		os.Exit(1)
	}

	// Close the listener when the application closes.
	defer l.Close()

	fmt.Println("Listening on " + HOST + ":" + PORT)
	for {
		// Listen for an incoming connection.
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting: ", err.Error())
			// TODO - should this be fatal?
			os.Exit(1)
		}
		// Handle connections in a new goroutine.
		go handleRequest(conn)
	}
}
