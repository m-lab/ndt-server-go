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
	"flag"
	"html/template"
	"log"
	"net"
	"net/http"
	"os"

	"github.com/gorilla/websocket"
	"github.com/m-lab/ndt-server-go/protocol"
	"github.com/m-lab/ndt-server-go/tests"
)

const (
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

var upgrader = websocket.Upgrader{}

func echo(w http.ResponseWriter, r *http.Request) {
	log.Printf("%+v\n", r)
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()
	for {
		mt, message, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}
		log.Printf("recv: %s", message)
		err = c.WriteMessage(mt, message)
		if err != nil {
			log.Println("write:", err)
			break
		}
	}
}

func home(w http.ResponseWriter, r *http.Request) {
	log.Println("ws://" + r.Host + "/echo")
	homeTemplate.Execute(w, "ws://"+r.Host+"/echo")
}

func main() {

	flag.Parse()
	http.HandleFunc("/ndt_protocol", echo)
	http.HandleFunc("/", home)
	http.ListenAndServe("localhost:3010", nil)

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

var homeTemplate = template.Must(template.New("").Parse(`
<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<script>  
window.addEventListener("load", function(evt) {
    var output = document.getElementById("output");
    var input = document.getElementById("input");
    var ws;
    var print = function(message) {
        var d = document.createElement("div");
        d.innerHTML = message;
        output.appendChild(d);
    };
    document.getElementById("open").onclick = function(evt) {
        if (ws) {
            return false;
        }
        ws = new WebSocket("{{.}}");
        ws.onopen = function(evt) {
            print("OPEN");
        }
        ws.onclose = function(evt) {
            print("CLOSE");
            ws = null;
        }
        ws.onmessage = function(evt) {
            print("RESPONSE: " + evt.data);
        }
        ws.onerror = function(evt) {
            print("ERROR: " + evt.data);
        }
        return false;
    };
    document.getElementById("send").onclick = function(evt) {
        if (!ws) {
            return false;
        }
        print("SEND: " + input.value);
        ws.send(input.value);
        return false;
    };
    document.getElementById("close").onclick = function(evt) {
        if (!ws) {
            return false;
        }
        ws.close();
        return false;
    };
});
</script>
</head>
<body>
<table>
<tr><td valign="top" width="50%">
<p>Click "Open" to create a connection to the server, 
"Send" to send a message to the server and "Close" to close the connection. 
You can change the message and send multiple times.
<p>
<form>
<button id="open">Open</button>
<button id="close">Close</button>
<p><input id="input" type="text" value="Hello world!">
<button id="send">Send</button>
</form>
</td><td valign="top" width="50%">
<div id="output"></div>
</td></tr></table>
</body>
</html>
`))
