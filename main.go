package main // import "kkn.fi/cmd/tcpproxy"

// http://pub.gajendra.net/src/trampoline.c

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path"
	"sync"
)

func usage() {
	fmt.Fprintf(os.Stderr, "usage: %v local:port remote:port\n", path.Base(os.Args[0]))
}

func main() {
	log.SetFlags(0)
	log.SetPrefix("tcpproxy: ")
	flag.Usage = usage
	if len(os.Args) != 3 {
		usage()
		os.Exit(2)
	}
	laddr := os.Args[1]
	raddr := os.Args[2]
	serve(laddr, raddr)
}

func serve(laddr, raddr string) {
	proxy, err := net.Listen("tcp", laddr)
	if err != nil {
		log.Fatal(err)
	}
	for {
		local, err := proxy.Accept()
		if err != nil {
			log.Print(err)
			os.Exit(1)
		}
		go func(local net.Conn) {
			remote, err := net.Dial("tcp", raddr)
			if err != nil {
				log.Print(err)
				if err := local.Close(); err != nil {
					log.Print(err)
				}
				return
			}
			var wg sync.WaitGroup
			wg.Add(2)
			go xfer(local, remote, &wg)
			go xfer(remote, local, &wg)
			wg.Wait()
			if err := remote.Close(); err != nil {
				log.Print(err)
			}
			if err := local.Close(); err != nil {
				log.Print(err)
			}
		}(local)
	}
}

func xfer(dst io.Writer, src io.Reader, wg *sync.WaitGroup) {
	defer wg.Done()
	if _, err := io.Copy(dst, src); err != nil {
		log.Print(err)
	}
}
