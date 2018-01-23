package protocol

import (
	"bufio"
	"github.com/m-lab/ndt-server-go/util"
	"encoding/binary"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net"
	"strconv"
)

/*
 __  __
|  \/  | ___  ___ ___  __ _  __ _  ___  ___
| |\/| |/ _ \/ __/ __|/ _` |/ _` |/ _ \/ __|
| |  | |  __/\__ \__ \ (_| | (_| |  __/\__ \
|_|  |_|\___||___/___/\__,_|\__, |\___||___/
                            |___/

	Message serialization and deserialization.
*/

func read_message_internal(cc net.Conn, reader io.Reader) (
                           byte, []byte, error) {

	// 1. read type

	type_buff := make([]byte, 1)
	_, err := util.IoReadFull(cc, reader, type_buff)
	if err != nil {
		return 0, nil, err
	}
	msg_type := type_buff[0]
	log.Printf("ndt: message type: %d", msg_type)

	// 2. read length

	len_buff := make([]byte, 2)
	_, err = util.IoReadFull(cc, reader, len_buff)
	if err != nil {
		return 0, nil, err
	}
	msg_length := binary.BigEndian.Uint16(len_buff)
	log.Printf("ndt: message length: %d", msg_length)

	// 3. read body

	msg_body := make([]byte, msg_length)
	_, err = util.IoReadFull(cc, reader, msg_body)
	if err != nil {
		return 0, nil, err
	}
	log.Printf("ndt: message body: '%s'\n", msg_body)

	return msg_type, msg_body, nil
}

type json_message_t struct {
	Msg string `json:"msg"`
}

// ReadJsonMessage reads a JSON encoded NDT message from |reader|. You need
// also to supply |cc| because we need to set the read deadline to make sure
// we do not hang forever. Returns a triple: message type, message body
// decoded from JSON, error that occurred.
func ReadJsonMessage(cc net.Conn, reader io.Reader) (byte, string, error) {
	msg_type, msg_buff, err := read_message_internal(cc, reader)
	if err != nil {
		return 0, "", err
	}
	s_msg := &json_message_t{}
	err = json.Unmarshal(msg_buff, &s_msg)
	if err != nil {
		return 0, "", err
	}
	if s_msg == nil {
		return 0, "", errors.New("ndt: received literal 'null'")
	}
	return msg_type, s_msg.Msg, nil
}

func write_message_internal(cc net.Conn, writer *bufio.Writer,
                            message_type byte, encoded_body []byte) error {

	log.Printf("ndt: write any message: type=%d\n", message_type)
	log.Printf("ndt: write any message: length=%d\n", len(encoded_body))
	log.Printf("ndt: write any message: body='%s'\n", string(encoded_body))

	// 1. write type

	err := util.IoWriteByte(cc, writer, message_type)
	if err != nil {
		return err
	}

	// 2. write length

	if len(encoded_body) > 65535 {
		return errors.New("ndt: encoded_body is too long")
	}
	encoded_len := make([]byte, 2)
	binary.BigEndian.PutUint16(encoded_len, uint16(len(encoded_body)))
	_, err = util.IoWrite(cc, writer, encoded_len)
	if err != nil {
		return err
	}

	// 3. write message body

	_, err = util.IoWrite(cc, writer, encoded_body)
	if err != nil {
		return err
	}
	return util.IoFlush(cc, writer)
}

// WriteJsonMessage encodes as JSON and writes on |writer| the NDT message
// with type |message_type| and body |message_body|. You need also to supply
// |cc| because we need to set the write deadline on it. Returns the error.
func WriteJsonMessage(cc net.Conn, writer *bufio.Writer,
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
	return write_message_internal(cc, writer, message_type, data)
}

// ExtendedLoginMessage contains the extended-login-message data.
type ExtendedLoginMessage struct {
	Msg      string `json:"msg"`
	TestsStr string `json:"tests"`
	Tests    int
}

// ReadExtendedLogin reads the extended loging message from |reader|. You also
// need to supply |cc| because we need to set the read deadline on it. This
// function returns a tuple: the extended-loging-message pointer and the error.
func ReadExtendedLogin(cc net.Conn, reader io.Reader) (
                       *ExtendedLoginMessage, error) {

	// Read ordinary message

	msg_type, msg_buff, err := read_message_internal(cc, reader)
	if err != nil {
		return nil, err
	}
	if msg_type != KvMsgExtendedLogin {
		return nil, errors.New("ndt: received invalid message")
	}

	// Process input as JSON message and validate its fields

	el_msg := &ExtendedLoginMessage{}
	err = json.Unmarshal(msg_buff, &el_msg)
	if err != nil {
		return nil, err
	}
	if el_msg == nil {
		return nil, errors.New("ndt: received literal 'null'")
	}
	log.Printf("ndt: client version: %s", el_msg.Msg)
	log.Printf("ndt: test suite: %s", el_msg.TestsStr)
	el_msg.Tests, err = strconv.Atoi(el_msg.TestsStr)
	if err != nil {
		return nil, err
	}
	log.Printf("ndt: test suite as int: %d", el_msg.Tests)
	if (el_msg.Tests & KvTestStatus) == 0 {
		return nil, errors.New("ndt: client does not support TEST_STATUS")
	}

	return el_msg, nil
}

// WriteRaWstring writes |str| on |writer|. You need also to pass |cc| because
// we need to set the write deadline. Returns the error.
func WriteRawString(cc net.Conn, writer *bufio.Writer, str string) error {
	log.Printf("ndt: write raw string: '%s'", str)
	_, err := util.IoWriteString(cc, writer, str)
	if err != nil {
		return err
	}
	return util.IoFlush(cc, writer)
}
