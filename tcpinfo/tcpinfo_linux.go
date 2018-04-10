package tcpinfo

import (
	"errors"
	"log"
	"net"
	"syscall"
	"unsafe"
)

// TCPInfo2 implements better way to get tcpinfo
func TCPInfo2(conn *net.TCPConn) (*syscall.TCPInfo, error) {
	file, err := conn.File()
	if err != nil {
		log.Println("error in getting file for the connection!")
		return nil, err
	}
	defer file.Close()
	fd := file.Fd()

	var info syscall.TCPInfo
	infoLen := uint32(syscall.SizeofTCPInfo)
	if _, _, e1 := syscall.Syscall6(syscall.SYS_GETSOCKOPT, fd, syscall.SOL_TCP,
		syscall.TCP_INFO, uintptr(unsafe.Pointer(&info)), uintptr(unsafe.Pointer(&infoLen)), 0); e1 != 0 {
		return nil, errors.New("Syscall error")
	}
	return &info, nil
}

// SetMSS uses syscall to set the MSS value on a connection.
func SetMSS(tcp *net.TCPListener, mss int) error {
	file, err := tcp.File()
	if err != nil {
		log.Println("error in getting file for the connection!")
		return err
	}
	defer file.Close()
	err = syscall.SetsockoptInt(int(file.Fd()), syscall.SOL_TCP, syscall.TCP_MAXSEG, mss)
	if err != nil {
		log.Println("error in setting MSS option on socket:", err)
		return err
	}
	return nil
}
