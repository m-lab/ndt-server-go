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

	"github.com/m-lab/ndt-server-go/protocol"
	"github.com/m-lab/ndt-server-go/tests"
)

const (
	HOST = "localhost"
	PORT = "3001"
	TYPE = "tcp"
)

func handleRequest(conn net.Conn) {
	// Close the connection when you're done with it.
	defer conn.Close()
	rdr := bufio.NewReader(conn)

	login, err := protocol.ReadLogin(rdr)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(*login)
	// Kickoff message
	conn.Write([]byte("123456 654321"))
	// Read the incoming connection into the buffer.
	// Send a response back to person contacting us.
	// Write(conn, 1, "{\"msg\":\"9977\"}")

	protocol.SendJSON(conn, 1, protocol.SimpleMsg{"0"})
	//Write(conn, 1, "{\"msg\":\"0\"}")
	protocol.SendJSON(conn, 2, protocol.SimpleMsg{"v3.8.1"})

	protocol.SendJSON(conn, 2, protocol.SimpleMsg{"1 2 4 8 32"})
	//conn.Write([]byte{1, 4, '9', '9', '7', '7'})
	//conn.Write([]byte{1, 1, '0'})

	mbox, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		fmt.Println("Error listening:", err.Error())
		os.Exit(1)
	}

	_, port, err := net.SplitHostPort(mbox.Addr().String())
	done := make(chan bool, 1)
	go tests.MiddleBox(mbox, done)

	protocol.SendJSON(conn, 3, protocol.SimpleMsg{port})

	err = protocol.Read(rdr)
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
