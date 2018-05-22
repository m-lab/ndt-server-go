package main

import (
	"fmt"
	"github.com/m-lab/ndt-server-go/protocol"
	"net"
	"os"
)

const (
	// HOST is the host to listen on
	HOST = "localhost"
	// PORT is the port to listen on
	PORT = "3001"
	// TYPE is the protocol to listen on
	TYPE = "tcp"
)

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
		go protocol.HandleControlConnection(conn)
	}
}
