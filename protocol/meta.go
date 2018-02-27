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

	err := WriteJsonMessage(rdwr, KvTestPrepare, "")
	if err != nil {
		return err
	}
	err = WriteJsonMessage(rdwr, KvTestStart, "")
	if err != nil {
		return err
	}

	// Read a sequence of TEST_MSGs from client

	for {
		msg_type, msg_body, err := ReadJsonMessage(rdwr)
		if err != nil {
			return err
		}
		if msg_type != KvTestMsg {
			return errors.New("ndt: expected TEST_MSG from client")
		}
		if msg_body == "" {
			break
		}
		log.Printf("ndt: metadata from client: %s", msg_body)
	}

	// Send empty TEST_FINALIZE to client

	return WriteJsonMessage(rdwr, KvTestFinalize, "")
}
