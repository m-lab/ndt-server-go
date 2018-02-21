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

// These are the bit codes for the individual NDT tests.
const (
	TestMid TestCode = 1 << iota
	TestC2S
	TestS2C
	TestSFW
	TestStatus
	TestMeta
)

// header is local, but fields must be exported for json Unmarshal
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

// ErrIllegalMessage is returned when a message header does
// not conform to the binary protocol, e.g. when WebSockets
// is used.
var ErrIllegalMessage = errors.New("illegal message header")

// ReadMessage reads a single message from rdr, and determines which
// protocol it required.
// TODO - use more robust reader from botticelli ??
// TODO - postpone multi-protocol handling to later versions of the code.
func ReadMessage(rdr io.Reader) (Message, error) {
	// TODO - any extra data from rdr may be lost!
	brdr := bufio.NewReader(rdr)
	peek, err := brdr.Peek(3)
	if err != nil {
		log.Println(err)
		return Message{}, err
	}
	// Binary protocol will have 2 or 11 in the first byte.  Otherwise
	// this is a websocket request.  Legacy server would accept both
	// connection types on 3001, but this implementation does not, at
	// least for now.
	if peek[0] > ProtocolExtended {
		// TODO
		// Probably best way to handle this is to create a new connection
		// to the websockets handler, and proxy everything from this
		// connection to the websockets connection.  A little less ugly
		// than the alternatives.
		for i := 0; i < 8; i++ {
			line, _ := brdr.ReadString('\n')
			log.Printf("%s", string(line))
		}
		return Message{}, ErrIllegalMessage
	}

	var hdr header
	err = binary.Read(brdr, binary.BigEndian, &hdr)
	if err != nil {
		log.Println(err)
		return Message{}, err
	}
	log.Println(hdr)
	content := make([]byte, hdr.Length)
	err = binary.Read(brdr, binary.BigEndian, content)
	if err != nil && err != io.EOF {
		log.Println(err)
		return Message{}, err
	}
	return Message{hdr, content}, nil
}

// struct is local, but fields must be exported
// for json Unmarshal.
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

// ErrUnknownProtocol is returned when protocol is not recognized.
var ErrUnknownProtocol = errors.New("unknown protocol")

// Codes used by NDT to indicate legacy or extended protocol
const (
	ProtocolLegacy   = 2
	ProtocolExtended = 11
)

// ReadLogin reads the initial login message.
func ReadLogin(rdr io.Reader) (Login, error) {
	msg, err := ReadMessage(rdr)
	if err != nil {
		return Login{}, err
	}

	log.Println(msg.Header)
	switch msg.Header.MsgType {
	case byte(ProtocolLegacy):
		// TODO Handle legacy, without json
		panic("Not implemented")

	case byte(ProtocolExtended):
		// Handle extended, with json
		lj := loginJSON{}
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
		// TODO maybe use ErrIllegalMessage
		return Login{}, ErrUnknownProtocol
	}
}

// SimpleMsg helps encoding json messages.
type SimpleMsg struct {
	Msg string `json:"msg,string"`
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
