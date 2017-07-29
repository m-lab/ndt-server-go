// +build linux

package tcpinfo

import (
	"errors"
	"log"
	"net"
	"os"
	"syscall"
	"unsafe"
)

// Alternate (better) way to get tcpinfo
func TCPInfo2(conn *net.TCPConn) (syscall.TCPInfo, error) {
	var info syscall.TCPInfo
	file, err := conn.File()
	if err != nil {
		log.Println("error in getting file for the connection!")
		return info, err
	}
	fd := file.Fd()

	infoLen := uint32(syscall.SizeofTCPInfo)
	if _, _, e1 := syscall.Syscall6(syscall.SYS_GETSOCKOPT, fd, syscall.SOL_TCP, syscall.TCP_INFO, uintptr(unsafe.Pointer(&info)), uintptr(unsafe.Pointer(&infoLen)), 0); e1 != 0 {
		return info, errors.New("Syscall error")
	}
	return info, nil
}

// SetMSS uses syscall to set the MSS value on a connection.
func SetMSS(tcp *net.TCPListener, mss int) {
	file, err := tcp.File()
	if err != nil {
		log.Println("error in getting file for the connection!")
		os.Exit(1)
	}
	err = syscall.SetsockoptInt(int(file.Fd()), syscall.SOL_TCP, syscall.TCP_MAXSEG, mss)
	file.Close()
	if err != nil {
		log.Println("error in setting MSS option on socket:", err)
		os.Exit(1)
	}
}
