package protocol

import (
	"bufio"
	"errors"
	"log"
)

/*
 __  __ _____ _____  _
|  \/  | ____|_   _|/ \
| |\/| |  _|   | | / _ \
| |  | | |___  | |/ ___ \
|_|  |_|_____| |_/_/   \_\

*/

// RunMetaTest runs the META test. |rdwr| is the buffered reader and writer
// for the connection. Returns the error.
func RunMetaTest(rdwr *bufio.ReadWriter) error {

	// Send empty TEST_PREPARE and TEST_START messages to the client

	err := SendSimpleMsg(rdwr.Writer, MsgTestPrepare, "")
	if err != nil {
		return err
	}
	err = SendSimpleMsg(rdwr.Writer, MsgTestStart, "")
	if err != nil {
		return err
	}

	// Read a sequence of TEST_MSGs from client

	for {
		msg, err := ReadMessageJson(rdwr.Reader)
		if err != nil {
			return err
		}
		msgType := msg.Header.MsgType
		msgBody := string(msg.Content)
		if msgType != MsgTest {
			return errors.New("ndt: expected TEST_MSG from client")
		}
		if msgBody == "" {
			break
		}
		log.Printf("ndt: metadata from client: %s", msgBody)
	}

	// Send empty TEST_FINALIZE to client

	return SendSimpleMsg(rdwr.Writer, MsgTestFinalize, "")
}
