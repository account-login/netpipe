package netpipe

import (
	"flag"
	"io"
	"log"
	"net"
	"os"
	"strings"
	"sync"
)

type clientParam struct {
	addr  string
	key   string
	rsync bool
}

func ClientMain() {
	// log
	log.SetFlags(log.Ldate | log.Lmicroseconds)
	log.SetOutput(os.Stderr)

	// param
	param := clientParam{}

	flag.StringVar(&param.addr, "addr", "", "server address")
	flag.StringVar(&param.key, "key", "", "encryption key")
	flag.BoolVar(&param.rsync, "rsync", false, "rsync client mode")
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

	// create encrypted writer
	writer := makeWriter(param.key+"client2server", conn)

	// write rsync cmd
	if param.rsync {
		// rsync shell cmd line
		// opening connection using:
		// ./netpipe-client -addr "1.1.1.1:50887" -- -l user 1.1.1.1 rsync --server --sender -vvlogDtprze.iLsf . "~/file"
		fs := flag.NewFlagSet("", flag.ContinueOnError)
		_ = fs.String("l", "", "user")
		_ = fs.Parse(flag.Args())
		args := fs.Args() // skip -l user

		if len(args) <= 4 || args[1] != "rsync" {
			log.Fatalf("[ERROR] bad rsync args: %v", args)
		}

		rsyncCmd := strings.Join(args[1:], " ") // skip ip 1.1.1.1
		log.Printf("rsync cmd: %v", rsyncCmd)

		_, err := writer.Write([]byte("exec " + rsyncCmd + "\n"))
		if err != nil {
			log.Fatalf("[ERROR] write rsync cmd: %v", err)
		}
	}

	wg := &sync.WaitGroup{}
	wg.Add(2)

	// stdin -> conn
	go func() {
		defer wg.Done()
		defer conn.CloseWrite()

		_, err := io.Copy(writer, os.Stdin)
		if err != nil {
			log.Printf("[ERROR] io.Copy(): %v", err)
		}
		log.Printf("stdin -> conn done")
	}()
	// conn -> stdout
	go func() {
		defer wg.Done()
		defer os.Stdout.Close()

		reader := makeReader(param.key+"server2client", conn)
		_, err := io.Copy(os.Stdout, reader)
		if err != nil {
			log.Printf("[ERROR] io.Copy(): %v", err)
		}
		log.Printf("conn -> stdout done")
		// rsync will not close stdout, causing stdin -> conn hang
		if param.rsync {
			os.Exit(0)
		}
	}()

	wg.Wait()
}
