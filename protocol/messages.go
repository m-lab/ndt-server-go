package protocol

import (
	"bufio"
	"encoding/binary"
	"encoding/json"
	"errors"
	"log"
	"strconv"
)

// TODO(gfr): decide where we want to put these files. We probably already
// have some hint from already existing pull requests (e.g. "start").

// TODO(bassosimone): refactor the code in here to use the message recv and
// send layer that was recently merged into master as #21.

/*
 __  __
|  \/  | ___  ___ ___  __ _  __ _  ___  ___
| |\/| |/ _ \/ __/ __|/ _` |/ _` |/ _ \/ __|
| |  | |  __/\__ \__ \ (_| | (_| |  __/\__ \
|_|  |_|\___||___/___/\__,_|\__, |\___||___/
                            |___/

	Message serialization and deserialization.
*/

type json_message_t struct {
	Msg string `json:"msg"`
}

func write_message_internal(rdwr *bufio.ReadWriter,
	message_type byte, encoded_body []byte) error {

	log.Printf("ndt: write any message: type=%d\n", message_type)
	log.Printf("ndt: write any message: length=%d\n", len(encoded_body))
	log.Printf("ndt: write any message: body='%s'\n", string(encoded_body))

	// 1. write type

	err := rdwr.Writer.WriteByte(message_type)
	if err != nil {
		return err
	}

	// 2. write length

	if len(encoded_body) > 65535 {
		return errors.New("ndt: encoded_body is too long")
	}
	encoded_len := make([]byte, 2)
	binary.BigEndian.PutUint16(encoded_len, uint16(len(encoded_body)))
	_, err = rdwr.Writer.Write(encoded_len)
	if err != nil {
		return err
	}

	// 3. write message body

	_, err = rdwr.Writer.Write(encoded_body)
	if err != nil {
		return err
	}
	return rdwr.Writer.Flush()
}

// WriteJsonMessage encodes as JSON and writes on |rdwr| the NDT message
// with type |message_type| and body |message_body|.
func WriteJsonMessage(rdwr *bufio.ReadWriter,
	message_type byte, message_body string) error {

	s_msg := &json_message_t{
		Msg: message_body,
	}
	log.Printf("ndt: sending standard message: type=%d", message_type)
	log.Printf("ndt: sending standard message: body='%s'", message_body)
	data, err := json.Marshal(s_msg)
	if err != nil {
		return err
	}
	return write_message_internal(rdwr, message_type, data)
}

// ExtendedLoginMessage contains the extended-login-message data.
type ExtendedLoginMessage struct {
	Msg      string `json:"msg"`
	TestsStr string `json:"tests"`
	Tests    TestCode
}

// ReadExtendedLogin reads the extended loging message from |reader|. You also
// need to supply |cc| because we need to set the read deadline on it. This
// function returns a tuple: the extended-loging-message pointer and the error.
func ReadExtendedLogin(rdwr *bufio.ReadWriter) (
	*ExtendedLoginMessage, error) {
	el_msg := &ExtendedLoginMessage{}
	login, err := ReadLogin(rdwr.Reader)
	if err != nil {
		return nil, err
	}
	if !login.IsExtended {
		panic("unexpected message type")
	}
	el_msg.Tests = login.Tests
	el_msg.TestsStr = strconv.Itoa(int(login.Tests))
	el_msg.Msg = login.Version
	return el_msg, nil
}
