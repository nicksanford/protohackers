package main

import (
	"fmt"
	"io"
	"net"
	"os"
	"time"

	"github.com/google/uuid"
)

var port = "8888"

func init() {
	if v := os.Getenv("PORT"); v != "" {
		port = v
	}
}

func main() {
	addr, err := net.ResolveTCPAddr("tcp", ":"+port)
	if err != nil {
		fmt.Println("Unable to resolve tcp address on port: " + port + " error: " + err.Error())
		os.Exit(1)
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		fmt.Println("Error listening on " + addr.String() + " error: " + err.Error())
		l.Close()
		os.Exit(1)
	}

	fmt.Println("Listening on port: " + port)
	for {
		conn, err := l.AcceptTCP()
		if err != nil {
			fmt.Println("Error accepting connection on port: " + port + " error: " + err.Error())
			l.Close()
			os.Exit(1)
		}
		go handleRequest(conn)
	}
}

func handleRequest(conn *net.TCPConn) {
	timeoutDuration := 5 * time.Second
	uuid := uuid.New().String()
	var msgCount int = 1
	var addr = conn.RemoteAddr().String()
	defer func() {
		fmt.Printf("From session: %s, addr: %s, msg: %d. CLOSING CONNECTION\n", uuid, addr, msgCount)
		conn.Close()
	}()

	req := make([]byte, 1)
	buf := make([]byte, 0)
	for {
		trace := fmt.Sprintf("From session: %s, addr: %s, msg %d: ", uuid, addr, msgCount)
		err := conn.SetDeadline(time.Now().Add(timeoutDuration))
		if err != nil {
			return
		}
		readLen, err := conn.Read(req)

		switch err {

		case nil:
			buf = append(buf, req[:readLen]...)

		case io.EOF:
			buf = append(buf, req[:readLen]...)
			flush(conn, &buf, trace)
      return

		default:
			fmt.Printf(trace+"Error reading: %s, readLen: %d\n", err.Error(), readLen)
			return
		}
		msgCount++
	}
}
func flush(conn *net.TCPConn, bufPtr *[]byte, trace string) {
  buf := *bufPtr
	bytesWritten := 0
	fmt.Printf(trace+"EOF. Flushing buffer of %d bytes before closing connection.\n", len(buf))
	for bytesWritten < len(buf) {
		writeLen, err := conn.Write(buf[bytesWritten:])
		bytesWritten += writeLen
		if err != nil {
			fmt.Printf(trace+"Error writing: %s, wrote: %d bytes out of %d total\n", err.Error(), bytesWritten, len(buf))
			return
		}
	}
}
