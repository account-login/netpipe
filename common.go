package netpipe

import (
	"io"
	"log"
	"net"
	"sync"
)

type TCPConn2Writer struct {
	conn *net.TCPConn
}

func (s TCPConn2Writer) Write(buf []byte) (int, error) {
	return s.conn.Write(buf)
}

func (s TCPConn2Writer) Close() error {
	return s.conn.CloseWrite()
}

func CopyAndClose(wg *sync.WaitGroup, writer io.WriteCloser, reader io.Reader) {
	defer wg.Done()
	defer writer.Close()

	_, err := io.Copy(writer, reader)
	if err != nil {
		log.Printf("[ERROR] io.Copy(): %v", err)
	}
}
