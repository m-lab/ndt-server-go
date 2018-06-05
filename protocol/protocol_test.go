// Part of ndt-server-go <https://github.com/m-lab/ndt-server-go>, which
// is free software under the Apache v2.0 License.

package protocol_test

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"log"
	"testing"

	"github.com/m-lab/ndt-server-go/protocol"
)

func init() {
	// Always prepend the filename and line number.
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func TestReadMessage(t *testing.T) {
	buf := bytes.NewBuffer(make([]byte, 0, 200))
	m := "{\"msg\": \"4.0.0.1\", \"tests\": \"63\"}"
	buf.Write([]byte{11, 0, byte(len(m))})
	buf.WriteString(m)

	msg, err := protocol.RecvBinaryMessage(bufio.NewReader(buf))
	if err != nil {
		t.Error(err.Error())
		return
	}
	if len(msg.Content) != 33 {
		t.Error("Wrong content length: ", len(msg.Content))
	}
}

func TestReadMany(t *testing.T) {
	buf := bytes.NewBuffer(make([]byte, 0, 200))
	m := "{\"msg\": \"4.0.0.1\", \"tests\": \"63\"}"
	buf.Write([]byte{11, 0, byte(len(m))})
	buf.WriteString(m)

	m = "{\"msg\": \"4.0.0\", \"tests\": \"6\"}"
	buf.Write([]byte{11, 0, byte(len(m))})
	buf.WriteString(m)

	reader := bufio.NewReader(buf)

	msg, err := protocol.RecvBinaryMessage(reader)
	if err != nil {
		t.Error(err.Error())
		return
	}
	if len(msg.Content) != 33 {
		t.Error("Wrong content length: ", len(msg.Content))
	}

	msg, err = protocol.RecvBinaryMessage(reader)
	if err != nil {
		t.Error(err.Error())
		return
	}
	if len(msg.Content) != 30 {
		t.Error("Wrong content length: ", len(msg.Content))
	}
}

func TestRecvLogin11(t *testing.T) {
	buf := bytes.NewBuffer(make([]byte, 0, 200))
	msg := "{\"msg\": \"4.0.0.1\", \"tests\": \"63\"}"
	buf.Write([]byte{11, 0, byte(len(msg))})
	buf.WriteString(msg)

	login, err := protocol.RecvLogin(bufio.NewReader(buf))
	if err != nil {
		t.Error(err.Error())
	}
	if !login.IsExtended {
		t.Error("IsExtended should be true")
	}
	if login.Version != "4.0.0.1" {
		t.Error("Version incorrectly parsed: ", login.Version)
	}
	if login.Tests != 63 {
		t.Error("Tests should be 63: ", login.Tests)
	}
}

func TestWriteMessage(t *testing.T) {
	outputBuf := bytes.NewBuffer(make([]byte, 0, 200))
	biow := bufio.NewWriter(outputBuf)
	err := protocol.SendBinaryMessage(biow, 4, []byte("abcdef"))
	if err != nil {
		t.Error(err.Error())
	}
	msgType, err := outputBuf.ReadByte()
	if err != nil {
		t.Error(err.Error())
	}
	if msgType != 4 {
		t.Error("unexpected message type: ", msgType)
	}
	var length int16 = 0
	err = binary.Read(outputBuf, binary.BigEndian, &length)
	if err != nil {
		t.Error(err.Error())
	}
	if length != 6 {
		t.Error("unexpected message length: ", length)
	}
	body := outputBuf.String()
	if body != "abcdef" {
		t.Error("unexpected message body: ", body)
	}
}
