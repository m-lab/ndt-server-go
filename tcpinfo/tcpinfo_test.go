package tcpinfo_test

import (
	"fmt"
	"log"
	"net"
	"os"
	"sync"
	"testing"

	"github.com/m-lab/ndt-server-go/tcpinfo"
)

func listen(port *net.TCPAddr, wg *sync.WaitGroup, ch chan net.Conn) {
	lnr, err := net.ListenTCP("tcp", port)
	defer lnr.Close()

	tcp, err := lnr.Accept()
	if err != nil {
		fmt.Println("Error accepting: ", err.Error())
		os.Exit(1)
	}

	ch <- tcp
	wg.Wait()
}

// Alternate (better) way to get tcpinfo
func BenchmarkTCPInfo2(b *testing.B) {

	serve, _ := net.ResolveTCPAddr("tcp", ":0")
	log.Println(serve.String())
	var wg sync.WaitGroup
	var ch chan net.Conn
	go listen(serve, &wg, ch)

	client, _ := net.ResolveTCPAddr("tcp", ":0")
	dialer, err := net.DialTCP("tcp", client, serve)
	if err != nil {
		fmt.Println("Error dialing: ", err.Error())
		os.Exit(1)
	}

	_ = <-ch
	for n := 0; n < b.N; n++ {
		tcpinfo.TCPInfo2(dialer)
	}

	wg.Done()
}
