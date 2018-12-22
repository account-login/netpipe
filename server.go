package netpipe

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/exec"
	"sync"
)

type serverParam struct {
	cmdline []string
	key     string
}

func handler(conn *net.TCPConn, param *serverParam) {
	defer conn.Close()

	log.Printf("%v -> %v accepted", conn.RemoteAddr(), conn.LocalAddr())

	// prepare pipe to cmd
	cmd := exec.Command(param.cmdline[0], param.cmdline[1:]...)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		log.Printf("[ERROR] cmd.StdinPipe(): %v", err)
		return
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Printf("[ERROR] cmd.StdoutPipe(): %v", err)
		return
	}
	cmd.Stderr = os.Stderr

	// start cmd
	err = cmd.Start()
	if err != nil {
		log.Printf("[ERROR] cmd.Start(): %v", err)
		return
	}

	wg := &sync.WaitGroup{}
	wg.Add(2)

	// cmd -> conn
	go func() {
		defer wg.Done()
		defer conn.CloseWrite()

		writer := makeWriter(param.key+"server2client", conn)
		_, err := io.Copy(writer, stdout)
		if err != nil {
			log.Printf("[ERROR] io.Copy(): %v", err)
		}
		log.Printf("cmd -> conn done")
	}()
	// conn -> cmd
	go func() {
		defer wg.Done()
		defer stdin.Close()

		reader := makeReader(param.key+"client2server", conn)
		_, err := io.Copy(stdin, reader)
		if err != nil {
			log.Printf("[ERROR] io.Copy(): %v", err)
		}
		log.Printf("conn -> cmd done")
	}()

	wg.Wait()
	log.Printf("%v -> %v leave", conn.RemoteAddr(), conn.LocalAddr())
}

func getPubIP() string {
	conn, err := net.Dial("tcp", "www.baidu.com:80")
	if err != nil {
		log.Printf("[ERROR] can not get public ip: %v", err)
		return "0.0.0.0"
	}

	conn.Close()
	laddr := conn.LocalAddr()
	return laddr.(*net.TCPAddr).IP.String()
}

func ServerMain() {
	// log
	log.SetFlags(log.Ldate | log.Lmicroseconds)
	log.SetOutput(os.Stderr)

	// parse flags
	flag.Parse()

	param := serverParam{}
	param.cmdline = flag.Args()
	if len(param.cmdline) == 0 {
		log.Fatal("empty cmdline")
	}

	// generate key
	param.key = genKey()

	// listen
	listener, err := net.ListenTCP("tcp", nil)
	if err != nil {
		log.Fatalf("[ERROR] listen: %v", err)
	}
	defer listener.Close()

	log.Printf("listen on: %v", listener.Addr())
	log.Printf("cmdline: %v", param.cmdline)

	// print client cmd
	pubIP := getPubIP()
	port := listener.Addr().(*net.TCPAddr).Port
	fmt.Printf("# rsync cmd\n")
	fmt.Printf("rsync -vvaz -e 'netpipe-client -rsync -addr %s:%d -key %s --'\n", pubIP, port, param.key)

	for {
		conn, err := listener.AcceptTCP()
		if err != nil {
			log.Printf("[ERROR] accept: %v", err)
			continue
		}

		go handler(conn, &param)
	}
}
