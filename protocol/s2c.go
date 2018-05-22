package protocol

import (
	"bufio"
	"encoding/json"
	"errors"
	"github.com/m-lab/ndt-server-go/netx"
	"github.com/m-lab/ndt-server-go/util"
	"log"
	"net"
	"strconv"
	"time"
)

/*
 ____ ____   ____
/ ___|___ \ / ___|
\___ \ __) | |
 ___) / __/| |___
|____/_____|\____|

*/

const kv_parallel_streams int = 3

type s2c_message_t struct {
	ThroughputValue  string
	UnsentDataAmount string
	TotalSentByte    string
}

// RunS2cTest runs the S2C test. |rdwr| is the buffered reader and writer
// for the connection. |is_extended| is true when we want to run a multi-stream
// test. Returns the error.
func RunS2cTest(rdwr *bufio.ReadWriter, is_extended bool) error {

	// Bind port and tell the port number to the server
	// TODO: choose a random port instead than an hardcoded port

	deadline := time.Now().Add(netx.DefaultTimeout)
	listener, err := netx.NewTCPListenerWithDeadline(":3010", deadline)
	if err != nil {
		return err
	}
	defer listener.Close()
	prepare_message := "3010"
	if is_extended {
		prepare_message += " 10000.0 1 500.0 0.0 "
		prepare_message += strconv.Itoa(kv_parallel_streams)
	}
	err = WriteJsonMessage(rdwr, MsgTestPrepare, prepare_message)
	if err != nil {
		return err
	}

	// Wait for client(s) to connect

	nstreams := 1
	if is_extended {
		nstreams = kv_parallel_streams
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

	err = WriteJsonMessage(rdwr, MsgTestStart, "")
	if err != nil {
		return err
	}

	// Run the N streams in parallel

	channel := make(chan int64)

	gen := util.NewBytesGenerator()
	output_buff := gen.GenLettersFast(8192)
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

			conn_writer := bufio.NewWriter(netx.NewDeadlineConn(conn))
			defer conn.Close()

			for {
				_, err = conn_writer.Write(output_buff)
				if err != nil {
					log.Println("ndt: failed to write to client")
					break
				}
				err = conn_writer.Flush()
				if err != nil {
					log.Println("ndt: cannot flush connection with client")
					break
				}
				channel <- int64(len(output_buff))
				if time.Since(start).Seconds() > 10.0 {
					log.Println("ndt: enough time elapsed")
					break
				}
			}

			conn.Close()  // Explicit to notify the client we're done
			channel <- -1 // Tell the controller we're done
		}(conns[idx])
	}

	bytes_sent := int64(0)
	for num_complete := 0; num_complete < len(conns); {
		count := <-channel
		if count < 0 {
			log.Printf("ndt: a stream just terminated...")
			num_complete += 1
			continue
		}
		bytes_sent += count
	}
	elapsed := time.Since(start)

	// Send message containing what we measured

	speed_kbits := (8.0 * float64(bytes_sent)) / 1000.0 / elapsed.Seconds()
	message := &s2c_message_t{
		ThroughputValue:  strconv.FormatFloat(speed_kbits, 'f', -1, 64),
		UnsentDataAmount: "0", // XXX
		TotalSentByte:    strconv.FormatInt(bytes_sent, 10),
	}
	data, err := json.Marshal(message)
	if err != nil {
		return err
	}
	err = write_message_internal(rdwr, MsgTest, data) // XXX
	if err != nil {
		return err
	}

	// Receive message from client containing its measured speed

	msg, err := ReadMessageJson(rdwr.Reader)
	if err != nil {
		return err
	}
	msgType := msg.Header.MsgType
	if msgType != MsgTest {
		return errors.New("ndt: received unexpected message from client")
	}
	msgBody := string(msg.Content)
	log.Printf("ndt: client measured speed: %s", msgBody)

	// FIXME: here we should send the web100 variables

	// Send the TEST_FINALIZE message that concludes the test

	return WriteJsonMessage(rdwr, MsgTestFinalize, "")
}
