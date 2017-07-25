package protocol_test

import (
	"bytes"
	"testing"

	"github.com/m-lab/ndt-server-go/protocol"
)

func TestReadLogin(t *testing.T) {
	buf := bytes.NewBuffer(make([]byte, 0, 200))
	msg := "{\"msg\": \"4.0.0.1\", \"tests\": \"63\"}"
	buf.Write([]byte{11, 0, byte(len(msg))})
	buf.WriteString(msg)

	login, err := protocol.ReadLogin(buf)
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
