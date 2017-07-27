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
	"io"
	"log"
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

// Message contains a header and arbitrary content.
type Message struct {
	// Header contains the message type and length
	Header header
	// Content contains the message body, which may be json or binary.
	Content []byte
}

// IllegalMessage is returned when a message header does
// not conform to the binary protocol, e.g. when WebSockets
// is used.
var IllegalMessage = errors.New("Illegal Message Header")

func ReadMessage(rdr io.Reader) (Message, error) {
	peeker := bufio.NewReader(rdr)
	get, err := peeker.Peek(3)
	if err != nil {
		log.Println(err)
		return Message{}, err
	}
	if get[0] > 11 {
		for i := 0; i < 8; i++ {
			line, _ := peeker.ReadString('\n')
			log.Printf("%s", string(line))
		}
		return Message{}, IllegalMessage

	}

	var hdr header
	err = binary.Read(rdr, binary.BigEndian, &hdr)
	if err != nil {
		log.Println(err)
		return Message{}, err
	}
	log.Println(hdr)
	content := make([]byte, hdr.Length)
	err = binary.Read(rdr, binary.BigEndian, content)
	if err != nil {
		log.Println(err)
		return Message{}, err
	}
	return Message{hdr, content}, nil
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

// ReadLogin reads the initial login message.
func ReadLogin(rdr io.Reader) (Login, error) {
	msg, err := ReadMessage(rdr)
	if err != nil {
		return Login{}, err
	}

	log.Println(msg.Header)
	switch msg.Header.MsgType {
	case byte(2):
		// TODO Handle legacy, without json
		panic("Not implemented")

	case byte(11):
		// Handle extended, with json
		lj := loginJSON{"foo", "bar"}
		err := json.Unmarshal(msg.Content, &lj)
		if err != nil {
			log.Println("Error: ", err)
		}
		tests, err := strconv.Atoi(lj.Tests)
		if err != nil {
			log.Println("Error: ", err)
		}
		return Login{byte(tests), lj.Msg, true}, nil
	default:
	}
	return Login{}, errors.New("Error")
}

// SimpleMsg helps encoding json messages.
type SimpleMsg struct {
	Msg string `json:"msg, string"`
}

// Send sends a raw message to the client.
func Send(conn io.Writer, t byte, msg []byte) {
	buf := make([]byte, 0, 3+len(msg))
	w := bytes.NewBuffer(buf)
	binary.Write(w, binary.BigEndian, t)
	binary.Write(w, binary.BigEndian, int16(len(msg)))
	binary.Write(w, binary.BigEndian, msg)
	w.WriteTo(conn)
}

// SendJSON sends a json encoded message to the client.
func SendJSON(conn io.Writer, t byte, msg interface{}) {
	j, err := json.Marshal(msg)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	Send(conn, t, j)
}
