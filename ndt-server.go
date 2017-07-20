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
	// HOST is the host to listen on
	HOST = "localhost"
	// PORT is the port to listen on
	PORT = "3001"
	// TYPE is the protocol to listen on
	TYPE    = "tcp"
	NDTPort = "3001"
)

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
	// protocol.SendJSON(conn, 2, protocol.SimpleMsg{"1 2 4 8 32"})
	protocol.SendJSON(conn, 2, protocol.SimpleMsg{"1"})

	tests.DoMiddleBox(conn)
	protocol.SendJSON(conn, 8, protocol.SimpleMsg{"Results 1"})
	protocol.SendJSON(conn, 8, protocol.SimpleMsg{"...Results 2"})
	protocol.Send(conn, 9, []byte{})
}

func main() {
	// TODO - does this listen on both ipv4 and ipv6?
	l, err := net.Listen("tcp", "localhost:"+NDTPort)
	if err != nil {
		fmt.Println("Error listening:", err.Error())
		os.Exit(1)
	}

	// Close the listener when the application closes.
	defer l.Close()

	fmt.Println("Listening on port " + NDTPort)
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
