// Part of ndt-server-go <https://github.com/m-lab/ndt-server-go>, which
// is free software under the Apache v2.0 License.

// Package protocol includes the NDT protocol elements
package protocol

/*
 * MSG_LOGIN uses binary protocol.
 * MSG_EXTENDED_LOGIN uses binary message types, but json message bodies.
 */

// TODO(bassosimone): import the message codes from botticelli.

import (
	"bufio"
	"encoding/binary"
	"encoding/json"
	"errors"
	"io"
	"log"
	"strconv"
)

// TestCode is used to decode the tests bitvector.
type TestCode int

const (
	// TestMid is the middle boxes test.
	TestMid TestCode = 1 << iota
	// TestC2S is the single-stream upload test.
	TestC2S
	// TestS2C is the single-stream download test.
	TestS2C
	// TestSFW is the simple firewall test.
	TestSFW
	// TestStatus indicates that we support waiting in queue.
	TestStatus
	// TestMeta indicate that we will send metadata.
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

// ErrIllegalMessageHeader is returned when a message header does not conform
// to the binary protocol, e.g. when WebSockets is used.
var ErrIllegalMessageHeader = errors.New("Illegal Message Header")

// ReadMessage reads a NDT message from |brdr|. Returns the message and/or the
// error that occurred while reading such message.
func ReadMessage(brdr *bufio.Reader) (Message, error) {
	// Implementation note: we use a buffered reader, so we're robust to
	// the case in which we receive a batch of messages.
	get, err := brdr.Peek(3)
	if err != nil {
		log.Println(err)
		return Message{}, err
	}
	if get[0] > 11 {
		// TODO
		// Probably best way to handle this is to create a new connection
		// to the websockets handler, and proxy everything from this
		// connection to the websockets connection.  A little less ugly
		// than the alternatives.
		for i := 0; i < 8; i++ {
			line, _ := brdr.ReadString('\n')
			log.Printf("%s", string(line))
		}
		return Message{}, ErrIllegalMessageHeader
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
	// TODO(bassosimone): decide whether we want to tolerate EOF (it
	// seems to me the original protocol does not).
	if err != nil && err != io.EOF {
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
func ReadLogin(brdr *bufio.Reader) (Login, error) {
	msg, err := ReadMessage(brdr)
	if err != nil {
		return Login{}, err
	}

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
		return Login{byte(tests), lj.Msg, true}, err
	default:
		// NOTHING
	}
	return Login{}, errors.New("Error")
}

// SimpleMsg helps encoding json messages.
type SimpleMsg struct {
	Msg string `json:"msg, string"`
}

// Send sends a raw message to the client.
func Send(wr *bufio.Writer, t byte, msg []byte) error {
	// Implementation note: here we could also use a net.Conn and a
	// buffer backed by a char slice. However, since we're coding the
	// protocol reader to be a *bufio.Reader, it would probably be
	// more handy to create initially a *bufio.ReadWriter and, then,
	// use that structure everywhere than having to pass around a
	// *bufio.Reader and a net.Conn.
	err := binary.Write(wr, binary.BigEndian, t)
	if err != nil {
		return err
	}
	err = binary.Write(wr, binary.BigEndian, int16(len(msg)))
	if err != nil {
		return err
	}
	err = binary.Write(wr, binary.BigEndian, msg)
	if err != nil {
		return err
	}
	return wr.Flush()
}

// SendJSON sends a json encoded message to the client.
func SendJSON(wr *bufio.Writer, t byte, msg interface{}) error {
	j, err := json.Marshal(msg)
	if err != nil {
		log.Println(err)
		return err
	}
	return Send(wr, t, j)
}
