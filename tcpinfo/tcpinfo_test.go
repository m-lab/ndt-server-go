package tcpinfo_test

import (
	"fmt"
	"log"
	"net"
	"os"
	"sync"
	"syscall"
	"testing"

	"github.com/m-lab/ndt-server-go/tcpinfo"
)

func listen(ln net.Listener, wg *sync.WaitGroup, ch chan net.Conn) {
	defer ln.Close()

	log.Println("Listening on ", ln.Addr())
	tcp, err := ln.Accept()
	if err != nil {
		fmt.Println("Error accepting: ", err.Error())
		os.Exit(1)
	}
	log.Println("Connected")

	ch <- tcp
	log.Println("Sent conn")
	wg.Wait()
}

func BenchmarkTCPInfo2(b *testing.B) {

	// listen on all interfaces
	ln, _ := net.Listen("tcp", ":0")

	var wg sync.WaitGroup
	wg.Add(1)
	ch := make(chan net.Conn)
	go listen(ln, &wg, ch)

	log.Println("Dialing")
	dialer, err := net.Dial("tcp", ln.Addr().String())
	defer dialer.Close()
	if err != nil {
		fmt.Println("Error dialing: ", err.Error())
		os.Exit(1)
	}

	log.Println("Waiting")
	tcp := <-ch
	log.Println("Connected on ", tcp.RemoteAddr())
	var info syscall.TCPInfo
	for n := 0; n < b.N; n++ {
		info, _ = tcpinfo.TCPInfo2(dialer.(*net.TCPConn))
	}

	wg.Done()
	log.Printf("%+v\n", info)
}
