package main

/*
MSG_LOGIN uses binary protocol
MSG_EXTENDED_LOGIN uses binary message types, but json message bodies.

Testing:
  websockets: 
	from ndt node_tests directory...
	   nodejs ndt_client.js --server localhost
	   (may need to 'npm install ws' in local directory)
  raw: 
    from ndt base directory
       src/web100clt -n localhost -dddddd -u `pwd` --enableprotolog
*/

import (
	"bufio"
	"log"
	"net"
	"os"

	"github.com/m-lab/ndt-server-go/protocol"
	"github.com/m-lab/ndt-server-go/tests"
)

const (
	// HOST is the host to listen on
	HOST = "localhost"
	// PORT is the port to listen on
	PORT = "3001"
	// TYPE is the protocol to listen on
	TYPE    = "tcp"
	NDTPort = "3001"
)

func init() {
	// Always prepend the filename and line number.
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func handleRequest(conn net.Conn) {
	// Close the connection when you're done with it.
	defer conn.Close()
	rdr := bufio.NewReader(conn)

	// Read the incoming login message.
	login, err := protocol.ReadLogin(rdr)
	if err != nil {
		log.Println(err)
		return
	}

	log.Println(login)
	// Send "Kickoff" message
	conn.Write([]byte("123456 654321"))
	// Send next messages in the handshake.
	protocol.SendJSON(conn, 1, protocol.SimpleMsg{"0"})
	protocol.SendJSON(conn, 2, protocol.SimpleMsg{"v3.8.1"})

	// TODO - this should be in response to the actual request.
	// protocol.SendJSON(conn, 2, protocol.SimpleMsg{"1 2 4 8 32"})
	protocol.SendJSON(conn, 2, protocol.SimpleMsg{"1"})

	tests.DoMiddleBox(conn)
	protocol.SendJSON(conn, 8, protocol.SimpleMsg{"Results 1"})
	protocol.SendJSON(conn, 8, protocol.SimpleMsg{"...Results 2"})
	protocol.Send(conn, 9, []byte{})
}

func main() {
	// TODO - does this listen on both ipv4 and ipv6?
	l, err := net.Listen("tcp", "localhost:"+NDTPort)
	if err != nil {
		log.Println("Error listening:", err.Error())
		os.Exit(1)
	}

	// Close the listener when the application closes.
	defer l.Close()

	log.Println("Listening on port " + NDTPort)
	for {
		// Listen for an incoming connection.
		conn, err := l.Accept()
		if err != nil {
			log.Println("Error accepting: ", err.Error())
			// TODO - should this be fatal?
			os.Exit(1)
		}
		// Handle connections in a new goroutine.
		go handleRequest(conn)
	}
}
