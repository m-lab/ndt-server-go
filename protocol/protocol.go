// Package protocol includes the NDT protocol elements
package protocol

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
	"io"
	"net"
	"os"
	"strconv"
)

// TestCode is used to decode the tests bitvector.
type TestCode int

const (
	// TestMid and other enums indicate the individual tests.
	TestMid TestCode = 1 << iota
	TestC2S
	TestS2C
	TestSFW
	TestStatus
	TestMeta
)

type header struct {
	MsgType byte // The message type
	Length  int16
}

type loginJSON struct {
	Msg   string
	Tests string
}

// Login represents a client login message.
type Login struct {
	Tests      byte   // The client test bits
	Version    string // The client version string
	IsExtended bool   // Type 11
}

// Read reads an unknown message type from the buffer.
func Read(rdr *bufio.Reader) error {
	fmt.Println("reading header")
	var hdr header
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

// ReadLogin reads the initial login message.
func ReadLogin(rdr io.Reader) (*Login, error) {
	fmt.Println("reading header")
	var hdr header
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
		lj := loginJSON{"foo", "bar"}
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

// Message contains a header and arbitrary content.
type Message struct {
	// Header contains the message type and length
	Header header
	// Content contains the message body, which may be json or binary.
	Content []byte
}

func (msg *Message) Read(rdr io.Reader) error {
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

// SimpleMsg helps encoding json messages.
type SimpleMsg struct {
	Msg string `json:"msg, string"`
}

// Send sends a raw message to the client.
func Send(conn net.Conn, t byte, msg []byte) {
	buf := make([]byte, 0, 3+len(msg))
	w := bytes.NewBuffer(buf)
	binary.Write(w, binary.BigEndian, t)
	binary.Write(w, binary.BigEndian, int16(len(msg)))
	binary.Write(w, binary.BigEndian, msg)
	w.WriteTo(conn)
}

// SendJSON sends a json encoded message to the client.
func SendJSON(conn net.Conn, t byte, msg interface{}) {
	j, err := json.Marshal(msg)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	Send(conn, t, j)
}
