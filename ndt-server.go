package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
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

type TestCode int

const (
	TestMid TestCode = 1 << iota
	TestC2S
	TestS2C
	TestSFW
	TestStatus
	TestMeta
)

type NDTMessage11 struct {
	Code   byte
	Length int16
	Tests  byte
}

func handleRequest(conn net.Conn) {
	defer conn.Close()
	rdr := bufio.NewReader(conn)

	var msg11 NDTMessage11
	binary.Read(rdr, binary.BigEndian, &msg11)

	fmt.Println(msg11)
	// Read the incoming connection into the buffer.
	count := 0
	buf := make([]byte, 1024)
	for {
		reqLen, err := conn.Read(buf)
		if (reqLen+count) > 3 && count < 4 {
			tests := buf[3-count]
			if tests&byte(TestMid) > 0 {
				fmt.Println("TestMid")
			}
			fmt.Printf("%b\n", buf[3-count])
		}
		if err != nil {
			fmt.Println("Error reading:", err.Error())
			return
		}
		fmt.Println(buf[:reqLen])

	}
	// Send a response back to person contacting us.
	conn.Write([]byte("Message received.\n"))
	// Close the connection when you're done with it.
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
