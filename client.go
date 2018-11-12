package netpipe

import (
	"flag"
	"log"
	"net"
	"os"
	"strings"
	"sync"
)

type clientParam struct {
	addr  string
	rsync bool
}

func ClientMain() {
	// log
	log.SetFlags(log.Ldate | log.Lmicroseconds)
	log.SetOutput(os.Stderr)

	// param
	param := clientParam{}

	flag.BoolVar(&param.rsync, "rsync", false, "rsync client mode")
	flag.StringVar(&param.addr, "addr", "", "server address")
	flag.Parse()

	// connect to server
	tcpAddr, err := net.ResolveTCPAddr("tcp", param.addr)
	if err != nil {
		log.Fatalf("[ERROR] resolve: %v", err)
	}

	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		log.Fatalf("[ERROR] dial: %v", err)
	}
	defer conn.Close()

	log.Printf("%v -> %v connected", conn.LocalAddr(), conn.RemoteAddr())

	// write rsync cmd
	if param.rsync {
		// rsync shell cmd line
		// opening connection using:
		// ./netpipe-client -addr "1.1.1.1:50887" -- -l user 1.1.1.1 rsync --server --sender -vvlogDtprze.iLsf . "~/file"
		args := flag.Args()
		if len(args) <= 4 {
			log.Fatalf("[ERROR] bad rsync args: %v", args)
		}

		rsyncCmd := strings.Join(args[3:], " ") // skip -l user 1.1.1.1
		log.Printf("rsync cmd: %v", rsyncCmd)

		_, err := conn.Write([]byte("exec " + rsyncCmd + "\n"))
		if err != nil {
			log.Fatalf("[ERROR] write rsync cmd: %v", err)
		}
	}

	wg := &sync.WaitGroup{}
	wg.Add(2)

	// stdin -> conn
	go func() {
		CopyAndClose(wg, TCPConn2Writer{conn}, os.Stdin)
		log.Printf("stdin -> conn done")
	}()
	// conn -> stdout
	go func() {
		CopyAndClose(wg, os.Stdout, conn)
		log.Printf("conn -> stdout done")
		// rsync will not close stdout, causing stdin -> conn hang
		if param.rsync {
			os.Exit(0)
		}
	}()

	wg.Wait()
}
