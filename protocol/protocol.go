// Part of ndt-server-go <https://github.com/m-lab/ndt-server-go>, which
// is free software under the Apache v2.0 License.

// Package protocol includes the NDT protocol elements
package protocol

import (
	"bufio"
	"encoding/binary"
	"encoding/json"
	"errors"
	"io"
	"log"
	"strconv"
)

// Spec: https://github.com/ndt-project/ndt/wiki/NDTProtocol

// TestCode is used to decode the tests bitvector.
type TestCode int

// TODO(bassosimone): we should probably also have a define for the
// type of message that we can receive rather than using `byte`

// Message types. Note: compared to the original specification, I have added
// the `Msg` prefix to all messages not having it for clarity. Also, the
// TEST_MSG define is mapped onto the MsgTest constant.
const (
	// MsgCommFailure indicates a communication link failure.
	MsgCommFailure = byte(iota)
	// MsgSrvQueue is used for queue management.
	MsgSrvQueue
	// MsgLogin is the legacy, binary protocol login message.
	MsgLogin
	// MsgTestPrepare is used to indicate test parameters.
	MsgTestPrepare
	// MsgTestStart is used to start a test.
	MsgTestStart
	// MsgTest is a message exchanged during a test.
	MsgTest
	// MsgTestFinalize is used to terminate a test.
	MsgTestFinalize
	// MsgError indicates an error during a test.
	MsgError
	// MsgResults contains the tests results.
	MsgResults
	// MsgLogout terminates a test session.
	MsgLogout
	// MsgWaiting tells a server that a client is alive.
	MsgWaiting
	// MsgExtendedLogin is the JSON-protocol login message.
	MsgExtendedLogin
)

// Test identifiers:
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
	// TestC2SExt is the multi stream upload test.
	TestC2SExt
	// TestS2CExt is the multi stream download test.
	TestS2CExt
)

// Queue states returned to client:
const (
	// SrvQueueTestStartsNow indicates that a test can start now.
	SrvQueueTestStartsNow = "0"
	// SrvQueueHeartbeat request client to tell us it's alive.
	SrvQueueHeartbeat = "9990"
	// SrvQueueServerFault indicates that the session must be terminated.
	SrvQueueServerFault = "9977"
	// SrvQueueServerBusy indicates that the server is busy.
	SrvQueueServerBusy = "9987"
	// SrvQueueServerBusy60s indicates that a server is busy for > 60 s.
	SrvQueueServerBusy60s = "9999"
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
	var hdr header
	err := binary.Read(brdr, binary.BigEndian, &hdr)
	if err != nil {
		log.Println(err)
		return Message{}, err
	}
	log.Printf("ndt: message type: %d", hdr.MsgType)
	log.Printf("ndt: message length: %d", hdr.Length)
	content := make([]byte, hdr.Length)
	// TODO(bassosimone): discuss with @gfr10598 whether here it makes
	// sense to use a BigEndian reader since we're reading bytes.
	err = binary.Read(brdr, binary.BigEndian, content)
	// TODO(bassosimone): decide whether we want to tolerate EOF (it
	// seems to me the original protocol does not).
	if err != nil && err != io.EOF {
		log.Println(err)
		return Message{}, err
	}
	// TODO(bassosimone): what is the Go idiomatic way of having a logger
	// with different logging levels (this is probably LOG_DEBUG)?
	log.Printf("ndt: message body: '%s'\n", content)
	return Message{hdr, content}, nil
}

type loginJSON struct {
	Msg   string
	Tests string
}

// Login represents a client login message.
type Login struct {
	Tests      TestCode   // The client test bits
	Version    string     // The client version string
	IsExtended bool       // Type MsgExtendedLogin
}

// ReadLogin reads the initial login message.
func ReadLogin(brdr *bufio.Reader) (Login, error) {
	msg, err := ReadMessage(brdr)
	if err != nil {
		return Login{}, err
	}

	switch msg.Header.MsgType {
	case MsgLogin:
		// TODO(bassosimone): Handle legacy, without json
		return Login{}, errors.New("not implemented")

	case MsgExtendedLogin: // Handle extended login, i.e. with JSON
		lj := loginJSON{"foo", "bar"}
		err := json.Unmarshal(msg.Content, &lj)
		if err != nil {
			log.Println("Error: ", err)
			return Login{}, err
		}
		// TODO(bassosimone): what is the idiomatic way for this? Do
		// we need to also check other parts of the code base?
		//
		// Or was that caused by the fact that in botticelli I was
		// passing to Unmarshal a pointer to a pointer?
		/*
			in botticelli:
		if lj == nil {
			return Login{}, errors.New("received literal null")
		}
		*/
		if lj.Msg == "foo" || lj.Tests == "bar" {
			return Login{}, errors.New("invalid message")
		}
		tests, err := strconv.Atoi(lj.Tests)
		if err != nil {
			log.Println("Error: ", err)
			return Login{}, err
		}
		log.Printf("ndt: client version: %s", lj.Msg)
		log.Printf("ndt: test suite: %s", lj.Tests)
		login := Login{TestCode(tests), lj.Msg, true}
		if (login.Tests & TestStatus) == 0 {
			return Login{}, errors.New("missing TEST_STATUS")
		}
		return login, nil

	default:
		// FALLTHROUGH
	}
	return Login{}, errors.New("unhandled message type")
}

// SimpleMsg helps encoding json messages.
type SimpleMsg struct {
	Msg string `json:"msg, string"`
}

func ReadMessageJson(brdr *bufio.Reader) (Message, error) {
	msg, err := ReadMessage(brdr)
	if err != nil {
		return Message{}, err
	}
	simple := &SimpleMsg{};
	err = json.Unmarshal(msg.Content, &simple)
	if err != nil {
		return Message{}, err
	}
	nmsg := Message{};
	nmsg.Header.MsgType = msg.Header.MsgType
	// TODO(bassosimone): what is golang equivalent of INT_MAX?
	if len(simple.Msg) > 0xffff {
		panic("unexpected maximum message length")
	}
	nmsg.Header.Length = int16(len(simple.Msg))
	// TODO(bassosimone): understand whether this is a string copy and
	// if we can avoid this copy by using other types.
	nmsg.Content = []byte(simple.Msg)
	return nmsg, nil
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
