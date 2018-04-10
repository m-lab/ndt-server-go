package protocol

import (
	"bufio"
	"errors"
	"github.com/m-lab/ndt-server-go/netx"
	"net"
	"log"
	"strconv"
	"sync"
	"time"
)

/*
	The control protocol.
*/

func update_queue_pos(rdwr *bufio.ReadWriter, position int) error {
	err := WriteJsonMessage(rdwr, MsgSrvQueue, strconv.Itoa(position))
	if err != nil {
		return errors.New("ndt: cannot write SRV_QUEUE message")
	}
	err = WriteJsonMessage(rdwr, MsgSrvQueue, SrvQueueHeartbeat)
	if err != nil {
		return errors.New("ndt: cannot write SRV_QUEUE heartbeat message")
	}
	msg_type, _, err := ReadJsonMessage(rdwr)
	if err != nil {
		return errors.New("ndt: cannot read MSG_WAITING message")
	}
	if msg_type != MsgWaiting {
		return errors.New("ndt: received unexpected message from client")
	}
	return nil
}

var kv_test_pending bool = false
var kv_test_pending_mutex sync.Mutex

var KvProduct string = "ndt-server-go" // XXX move / change

// HandleControlConnection handles the control connection |cc|.
func HandleControlConnection(cc net.Conn) {
	cc = netx.NewDeadlineConn(cc)
	defer cc.Close()

	rdwr := bufio.NewReadWriter(bufio.NewReader(cc), bufio.NewWriter(cc))

	// Read extended login message

	login_msg, err := ReadExtendedLogin(rdwr)
	if err != nil {
		log.Println("ndt: cannot read extended login")
		return
	}

	// Write kickoff message

	err = WriteRawString(rdwr, "123456 654321")
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
		err = update_queue_pos(rdwr, 1)
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

	err = WriteJsonMessage(rdwr, MsgSrvQueue, "0")
	if err != nil {
		log.Println("ndt: cannot write SRV_QUEUE message")
		return
	}

	// Write server version to client

	err = WriteJsonMessage(rdwr, MsgLogin,
			"v3.7.0 (" + KvProduct + ")")
	if err != nil {
		log.Println("ndt: cannot send our version to client")
		return
	}

	// Send list of encoded tests IDs

	status := login_msg.Tests
	tests_message := ""
	if (status & TestS2CExt) != 0 {
		tests_message += strconv.Itoa(TestS2CExt)
		tests_message += " "
	}
	if (status & TestS2C) != 0 {
		tests_message += strconv.Itoa(TestS2C)
		tests_message += " "
	}
	if (status & TestMeta) != 0 {
		tests_message += strconv.Itoa(TestMeta)
	}
	err = WriteJsonMessage(rdwr, MsgLogin, tests_message)
	if err != nil {
		log.Println("ndt: cannot send the list of tests to client")
		return
	}

	// Run tests

	if (status & TestS2CExt) != 0 {
		err = RunS2cTest(rdwr, true)
		if err != nil {
			log.Println("ndt: failure to run s2c_ext test")
			return
		}
	}
	if (status & TestS2C) != 0 {
		err = RunS2cTest(rdwr, false)
		if err != nil {
			log.Println("ndt: failure running s2c test")
			return
		}
	}
	if (status & TestMeta) != 0 {
		err = RunMetaTest(rdwr)
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
	err = WriteJsonMessage(rdwr, MsgResults,
		"botticelli_does_not_yet_collect_web100_data_sorry: 1\n")
	if err != nil {
		return
	}

	// Send empty MSG_LOGOUT to client

	err = WriteJsonMessage(rdwr, MsgLogout, "")
	if err != nil {
		return
	}
}
