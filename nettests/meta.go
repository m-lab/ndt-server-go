package nettests

import (
	"bufio"
	"errors"
	"github.com/m-lab/ndt-server-go/protocol"
	"log"
)

// RunMetaTest runs the META test. |rdwr| is the buffered reader and writer
// for the connection. Returns the error.
func RunMetaTest(rdwr *bufio.ReadWriter) error {

	// Send empty TEST_PREPARE and TEST_START messages to the client

	err := protocol.SendSimpleMsg(rdwr.Writer, protocol.MsgTestPrepare, "")
	if err != nil {
		return err
	}
	err = protocol.SendSimpleMsg(rdwr.Writer, protocol.MsgTestStart, "")
	if err != nil {
		return err
	}

	// Read a sequence of TEST_MSGs from client

	for {
		msg, err := protocol.ReadMessageJson(rdwr.Reader)
		if err != nil {
			return err
		}
		msgType := msg.Header.MsgType
		msgBody := string(msg.Content)
		if msgType != protocol.MsgTest {
			return errors.New("ndt: expected TEST_MSG from client")
		}
		if msgBody == "" {
			break
		}
		log.Printf("ndt: metadata from client: %s", msgBody)
	}

	// Send empty TEST_FINALIZE to client

	return protocol.SendSimpleMsg(rdwr.Writer, protocol.MsgTestFinalize, "")
}
