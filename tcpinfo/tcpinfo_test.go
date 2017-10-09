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

func listen(ln net.Listener, wg *sync.WaitGroup) {
	defer ln.Close()

	log.Println("Listening on ", ln.Addr())
	_, err := ln.Accept()
	if err != nil {
		fmt.Println("Error accepting: ", err.Error())
		os.Exit(1)
	}

	wg.Wait()
}

func BenchmarkTCPInfo2(b *testing.B) {
	ln, _ := net.Listen("tcp", ":0")

	var wg sync.WaitGroup
	wg.Add(1)
	go listen(ln, &wg)

	dialer, err := net.Dial("tcp", ln.Addr().String())
	defer dialer.Close()
	if err != nil {
		fmt.Println("Error dialing: ", err.Error())
		os.Exit(1)
	}

	var info syscall.TCPInfo
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		info, _ = tcpinfo.TCPInfo2(dialer.(*net.TCPConn))
	}

	wg.Done()
	log.Printf("%+v\n", info)
}

func TestBasic(t *testing.T) {
	ln, _ := net.Listen("tcp", ":0")

	var wg sync.WaitGroup
	wg.Add(1)
	go listen(ln, &wg)

	dialer, err := net.Dial("tcp", ln.Addr().String())
	defer dialer.Close()
	if err != nil {
		fmt.Println("Error dialing: ", err.Error())
		os.Exit(1)
	}

	var info syscall.TCPInfo
	info, _ = tcpinfo.TCPInfo2(dialer.(*net.TCPConn))
	log.Printf("%+v\n", info)

}