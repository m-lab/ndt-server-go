package main

/*
MSG_LOGIN uses binary protocol
MSG_EXTENDED_LOGIN uses binary message types, but json message bodies.


*/

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"
)

const (
	HOST = "localhost"
	PORT = "3001"
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

type NDTHeader struct {
	MsgType byte // The message type
	Length  int16
}

type LoginJSON struct {
	Msg   string
	Tests string
}

type Login struct {
	tests      byte   // The client test bits
	version    string // The client version string
	isExtended bool   // Type 11
}

func Read(rdr *bufio.Reader) error {
	fmt.Println("reading header")
	var hdr NDTHeader
	err := binary.Read(rdr, binary.BigEndian, &hdr)
	if err != nil {
		return err
	}
	fmt.Println(hdr)
	content := make([]byte, hdr.Length)
	fmt.Println("reading body")
	err = binary.Read(rdr, binary.BigEndian, content)
	if err != nil {
		return err
	}
	fmt.Println(string(content))
	return nil
}

// Hardly seems worthwhile returning a pointer.  Options?
func ReadLogin(rdr *bufio.Reader) (*Login, error) {
	fmt.Println("reading header")
	var hdr NDTHeader
	err := binary.Read(rdr, binary.BigEndian, &hdr)
	if err != nil {
		return nil, err
	}
	content := make([]byte, hdr.Length)
	fmt.Println("reading body")
	err = binary.Read(rdr, binary.BigEndian, content)
	if err != nil {
		return nil, err
	}
	switch hdr.MsgType {
	case byte(2):
		os.Exit(11)
	// Handle legacy, without json

	case byte(11):
		// Handle extended, with json
		lj := LoginJSON{"foo", "bar"}
		err := json.Unmarshal(content, &lj)
		if err != nil {
			fmt.Println("Error: ", err)
		}
		tests, err := strconv.Atoi(lj.Tests)
		if err != nil {
			fmt.Println("Error: ", err)
		}
		return &Login{byte(tests), lj.Msg, true}, nil
	default:
	}
	return nil, errors.New("Error")
}

type NDTMessage struct {
	// Header contains the message type and length
	Header NDTHeader
	// Content contains the message body, which may be json or binary.
	Content []byte
}

func (msg *NDTMessage) Read(rdr *bufio.Reader) error {
	fmt.Println("reading header")
	err := binary.Read(rdr, binary.BigEndian, &msg.Header)
	if err != nil {
		return err
	}
	msg.Content = make([]byte, msg.Header.Length)
	fmt.Println("reading body")
	err = binary.Read(rdr, binary.BigEndian, msg.Content)
	if err != nil {
		return err
	}
	return nil
}

type SimpleMsg struct {
	Msg string `json:"msg, string"`
}

// Request is the struct received in the client MSG_EXTENDED_LOGIN
type Request struct {
	msg   string
	tests string
}

func Send(conn net.Conn, t byte, msg []byte) {
	buf := make([]byte, 0, 3+len(msg))
	w := bytes.NewBuffer(buf)
	binary.Write(w, binary.BigEndian, t)
	binary.Write(w, binary.BigEndian, int16(len(msg)))
	binary.Write(w, binary.BigEndian, msg)
	w.WriteTo(conn)
}

func SendJSON(conn net.Conn, t byte, msg interface{}) {
	j, err := json.Marshal(msg)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	Send(conn, t, j)
}

func handleRequest(conn net.Conn) {
	// Close the connection when you're done with it.
	defer conn.Close()
	rdr := bufio.NewReader(conn)

	//	var msg NDTMessage
	//	err := msg.Read(rdr)
	login, err := ReadLogin(rdr)
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

	SendJSON(conn, 1, SimpleMsg{"0"})
	//Write(conn, 1, "{\"msg\":\"0\"}")
	SendJSON(conn, 2, SimpleMsg{"v3.8.1"})

	SendJSON(conn, 2, SimpleMsg{"1 2 4 8 32"})
	//conn.Write([]byte{1, 4, '9', '9', '7', '7'})
	//conn.Write([]byte{1, 1, '0'})

	SendJSON(conn, 3, SimpleMsg{"12345"})

	err = Read(rdr)
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
