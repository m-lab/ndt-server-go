package control

import (
	"bufio"
	"github.com/m-lab/ndt-server-go/nettests"
	"github.com/m-lab/ndt-server-go/netx"
	"github.com/m-lab/ndt-server-go/protocol"
	"log"
	"net"
	"strconv"
	"sync"
)

var testsRunning int = 0
var testsRunningMutex sync.Mutex

const maxTestsRunning int = 32
const serverName string = "ndt-server-go"

// HandleControlConnection handles the control connection |cc|.
func HandleControlConnection(cc net.Conn) {
	cc = netx.NewDeadlineConn(cc)
	defer cc.Close()

	rdwr := bufio.NewReadWriter(bufio.NewReader(cc), bufio.NewWriter(cc))

	// Read extended login message
	login, err := protocol.ReadLogin(rdwr.Reader)
	if err != nil || login.IsExtended == false {
		log.Println("ndt: cannot read extended login message")
		return
	}

	// Write kickoff message
	kickoff := "123456 654321"
	count, err := rdwr.Writer.WriteString(kickoff)
	if err != nil || count != len(kickoff) {
		log.Println("ndt: cannot write kickoff message")
		return
	}
	err = rdwr.Writer.Flush()
	if err != nil {
		log.Println("ndt: cannot flush kickoff message")
		return
	}

	// Queue management (simplified)
	canContinue := false
	testsRunningMutex.Lock()
	if testsRunning < maxTestsRunning {
		testsRunning += 1
		canContinue = true
	}
	testsRunningMutex.Unlock()
	if !canContinue {
		log.Println("ndt: too many running tests")
		serverBusy := "9988"
		protocol.SendSimpleMsg(rdwr.Writer, protocol.MsgSrvQueue, serverBusy)
		return
	}

	log.Println("ndt: this test is now running")
	defer func() {
		log.Println("ndt: test complete; allowing another test to run")
		testsRunningMutex.Lock()
		testsRunning -= 1
		testsRunningMutex.Unlock()
	}()

	// Write queue empty message
	err = protocol.SendSimpleMsg(rdwr.Writer, protocol.MsgSrvQueue, "0")
	if err != nil {
		log.Println("ndt: cannot write SRV_QUEUE message")
		return
	}

	// Write server version to client
	err = protocol.SendSimpleMsg(rdwr.Writer, protocol.MsgLogin,
		"v3.7.0 ("+serverName+")")
	if err != nil {
		log.Println("ndt: cannot send our version to client")
		return
	}

	// Send list of encoded tests IDs
	status := login.Tests
	testsMessage := ""
	if (status & protocol.TestS2CExt) != 0 {
		testsMessage += strconv.Itoa(int(protocol.TestS2CExt))
		testsMessage += " "
	}
	if (status & protocol.TestS2C) != 0 {
		testsMessage += strconv.Itoa(int(protocol.TestS2C))
		testsMessage += " "
	}
	if (status & protocol.TestMeta) != 0 {
		testsMessage += strconv.Itoa(int(protocol.TestMeta))
	}
	err = protocol.SendSimpleMsg(rdwr.Writer, protocol.MsgLogin, testsMessage)
	if err != nil {
		log.Println("ndt: cannot send the list of tests to client")
		return
	}

	// Run tests
	if (status & protocol.TestS2CExt) != 0 {
		err = nettests.RunS2CTest(rdwr, true)
		if err != nil {
			log.Println("ndt: failure to run s2c_ext test")
			return
		}
	}
	if (status & protocol.TestS2C) != 0 {
		err = nettests.RunS2CTest(rdwr, false)
		if err != nil {
			log.Println("ndt: failure running s2c test")
			return
		}
	}
	if (status & protocol.TestMeta) != 0 {
		err = nettests.RunMetaTest(rdwr)
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
	err = protocol.SendSimpleMsg(rdwr.Writer, protocol.MsgResults,
		"botticelli_does_not_yet_collect_web100_data_sorry: 1\n")
	if err != nil {
		return
	}

	// Send empty MSG_LOGOUT to client
	err = protocol.SendSimpleMsg(rdwr.Writer, protocol.MsgLogout, "")
	if err != nil {
		return
	}
}
