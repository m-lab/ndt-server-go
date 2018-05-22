package nettests

import (
	"bufio"
	"encoding/json"
	"errors"
	"github.com/m-lab/ndt-server-go/netx"
	"github.com/m-lab/ndt-server-go/protocol"
	"github.com/m-lab/ndt-server-go/util"
	"log"
	"net"
	"strconv"
	"time"
)

const numParallelStreams int = 3

type s2cMessage struct {
	ThroughputValue  string
	UnsentDataAmount string
	TotalSentByte    string
}

// RunS2CTest runs the S2C test. |rdwr| is the buffered reader and writer
// for the connection. |isExtended| is true when we want to run a multi-stream
// test. Returns the error.
func RunS2CTest(rdwr *bufio.ReadWriter, isExtended bool) error {

	// Bind port and tell the port number to the client
	// TODO: choose a random port instead than an hardcoded port

	deadline := time.Now().Add(netx.DefaultTimeout)
	listener, err := netx.NewTCPListenerWithDeadline(":3010", deadline)
	if err != nil {
		return err
	}
	defer listener.Close()
	prepareMessage := "3010"
	if isExtended {
		prepareMessage += " 10000.0 1 500.0 0.0 "
		prepareMessage += strconv.Itoa(numParallelStreams)
	}
	err = protocol.SendSimpleMsg(rdwr.Writer, protocol.MsgTestPrepare, prepareMessage)
	if err != nil {
		return err
	}

	// Wait for client(s) to connect

	nstreams := 1
	if isExtended {
		nstreams = numParallelStreams
	}

	conns := make([]net.Conn, nstreams)
	for idx := 0; idx < len(conns); idx += 1 {
		conn, err := listener.Accept()
		if err != nil {
			return err
		}
		conns[idx] = conn
	}

	// Send empty TEST_START message to tell the client to start

	err = protocol.SendSimpleMsg(rdwr.Writer, protocol.MsgTestStart, "")
	if err != nil {
		return err
	}

	// Run the N streams in parallel

	channel := make(chan int64)

	gen := util.NewBytesGenerator()
	outputBuff := gen.GenLettersFast(8192)
	start := time.Now()

	for idx := 0; idx < len(conns); idx += 1 {
		log.Printf("ndt: start stream with id %d\n", idx)

		// Note: rather than creating and destroying the goroutine
		// always it would be more considerate to just have a few
		// already active goroutines to which to dispatch the message
		// that there is a specific connection to be served

		go func(conn net.Conn) {
			// Send the buffer to the client for about ten seconds
			// TODO: here we should take `web100` snapshots

			writer := bufio.NewWriter(netx.NewDeadlineConn(conn))
			defer conn.Close()

			for {
				amt, err := writer.Write(outputBuff)
				if err != nil || amt != len(outputBuff) {
					log.Println("ndt: failed to write to client")
					break
				}
				err = writer.Flush()
				if err != nil {
					log.Println("ndt: cannot flush connection with client")
					break
				}
				channel <- int64(len(outputBuff))
				if time.Since(start).Seconds() > 10.0 {
					log.Println("ndt: enough time elapsed")
					break
				}
			}

			conn.Close()  // Explicit to notify the client we're done
			channel <- -1 // Tell the controller we're done
		}(conns[idx])
	}

	bytesSent := int64(0)
	for numComplete := 0; numComplete < len(conns); {
		count := <-channel
		if count < 0 {
			log.Printf("ndt: a stream just terminated...")
			numComplete += 1
			continue
		}
		bytesSent += count
	}
	elapsed := time.Since(start)

	// Send message containing what we measured

	speedKbits := (8.0 * float64(bytesSent)) / 1000.0 / elapsed.Seconds()
	message := &s2cMessage{
		ThroughputValue:  strconv.FormatFloat(speedKbits, 'f', -1, 64),
		UnsentDataAmount: "0", // XXX
		TotalSentByte:    strconv.FormatInt(bytesSent, 10),
	}
	data, err := json.Marshal(message)
	if err != nil {
		return err
	}
	err = protocol.Send(rdwr.Writer, protocol.MsgTest, data)
	if err != nil {
		return err
	}

	// Receive message from client containing its measured speed

	msg, err := protocol.ReadMessageJson(rdwr.Reader)
	if err != nil {
		return err
	}
	msgType := msg.Header.MsgType
	if msgType != protocol.MsgTest {
		return errors.New("ndt: received unexpected message from client")
	}
	msgBody := string(msg.Content)
	log.Printf("ndt: client measured speed: %s", msgBody)

	// TODO(bassosimone): here we should send the web100 variables. The code
	// on the client side should work anyway, however.

	// Send the TEST_FINALIZE message that concludes the test

	return protocol.SendSimpleMsg(rdwr.Writer, protocol.MsgTestFinalize, "")
}
