package protocol

import (
	"bufio"
	"errors"
	"net"
	"log"
	"strconv"
	"sync"
	"time"
)

/*
	The control protocol.
*/

func update_queue_pos(cc net.Conn, reader *bufio.Reader, writer *bufio.Writer,
                      position int) error {
	err := WriteJsonMessage(cc, writer, KvSrvQueue,
			strconv.Itoa(position))
	if err != nil {
		return errors.New("ndt: cannot write SRV_QUEUE message")
	}
	err = WriteJsonMessage(cc, writer, KvSrvQueue,
			KvSrvQueueHeartbeat)
	if err != nil {
		return errors.New("ndt: cannot write SRV_QUEUE heartbeat message")
	}
	msg_type, _, err := ReadJsonMessage(cc, reader)
	if err != nil {
		return errors.New("ndt: cannot read MSG_WAITING message")
	}
	if msg_type != KvMsgWaiting {
		return errors.New("ndt: received unexpected message from client")
	}
	return nil
}

var kv_test_pending bool = false
var kv_test_pending_mutex sync.Mutex

var KvProduct string = "ndt-server-go" // XXX move / change

// HandleControlConnection handles the control connection |cc|.
func HandleControlConnection(cc net.Conn) {
	defer cc.Close()

	reader := bufio.NewReader(cc)
	writer := bufio.NewWriter(cc)

	// Read extended login message

	login_msg, err := ReadExtendedLogin(cc, reader)
	if err != nil {
		log.Println("ndt: cannot read extended login")
		return
	}

	// Write kickoff message

	err = WriteRawString(cc, writer, "123456 654321")
	if err != nil {
		log.Println("ndt: cannot write kickoff message")
		return
	}

	// Queue management
	// XXX The current implementation of queue management is minimal, and
	// possibly also very ugly and stupid. Must be improved.
	//
	// Moreover the lock/unlock dance with the mutex is not idiomatic
	// golang and it would be better to use messages and channels.

	for {
		kv_test_pending_mutex.Lock()
		if !kv_test_pending {
			kv_test_pending = true
			kv_test_pending_mutex.Unlock()
			break
		}
		kv_test_pending_mutex.Unlock()
		err = update_queue_pos(cc, reader, writer, 1)
		if err != nil {
			log.Println("ndt: failed to update client of its queue position")
			return
		}
		time.Sleep(3.0 * time.Second)
	}
	log.Println("ndt: this test is now running")
	defer func() {
		log.Println("ndt: test complete; allowing another test to run")
		kv_test_pending_mutex.Lock()
		kv_test_pending = false
		kv_test_pending_mutex.Unlock()
	}()

	// Write queue empty message

	err = WriteJsonMessage(cc, writer, KvSrvQueue, "0")
	if err != nil {
		log.Println("ndt: cannot write SRV_QUEUE message")
		return
	}

	// Write server version to client

	err = WriteJsonMessage(cc, writer, KvMsgLogin,
			"v3.7.0 (" + KvProduct + ")")
	if err != nil {
		log.Println("ndt: cannot send our version to client")
		return
	}

	// Send list of encoded tests IDs

	status := login_msg.Tests
	tests_message := ""
	if (status & KvTestS2cExt) != 0 {
		tests_message += strconv.Itoa(KvTestS2cExt)
		tests_message += " "
	}
	if (status & KvTestS2c) != 0 {
		tests_message += strconv.Itoa(KvTestS2c)
		tests_message += " "
	}
	if (status & KvTestMeta) != 0 {
		tests_message += strconv.Itoa(KvTestMeta)
	}
	err = WriteJsonMessage(cc, writer, KvMsgLogin, tests_message)
	if err != nil {
		log.Println("ndt: cannot send the list of tests to client")
		return
	}

	// Run tests

	if (status & KvTestS2cExt) != 0 {
		err = RunS2cTest(cc, reader, writer, true)
		if err != nil {
			log.Println("ndt: failure to run s2c_ext test")
			return
		}
	}
	if (status & KvTestS2c) != 0 {
		err = RunS2cTest(cc, reader, writer, false)
		if err != nil {
			log.Println("ndt: failure running s2c test")
			return
		}
	}
	if (status & KvTestMeta) != 0 {
		err = RunMetaTest(cc, reader, writer)
		if err != nil {
			log.Println("ndt: failure running meta test")
			return
		}
	}

	// Send MSG_RESULTS to the client

	/*
	 * TODO: Here we should actually send results but to do that we need
	 * first to implement reading Web100 variables from /proc/web100.
	 *
	 * Until we reach this point, send back a variable that NDT client
	 * will ignore but that is consistent with what it would expect.
	 */
	err = WriteJsonMessage(cc, writer, KvMsgResults,
		"botticelli_does_not_yet_collect_web100_data_sorry: 1\n")
	if err != nil {
		return
	}

	// Send empty MSG_LOGOUT to client

	err = WriteJsonMessage(cc, writer, KvMsgLogout, "")
	if err != nil {
		return
	}
}
