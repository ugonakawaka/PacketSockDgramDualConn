package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strings"

	dualconn "github.com/ugonakawaka/PacketSockDgramDualConn/packetsockdgramdualconn"
)

var (
	// AppName application name
	AppName = flag.String("name", "TEST UDP SERVER", "application name")
	// Port port
	Port = flag.Int("port", 55501, "message receive port")
	// Bufsize buffer size
	Bufsize = flag.Int("bsize", 1500, "recv buffer size")
)

func main() {
	flag.Parse()
	fmt.Println(*AppName, *Port, *Bufsize)

	ctx, cancel := context.WithCancel(context.Background())

	// udp handler
	handler := func(n int, iph *dualconn.IpHeader, udph *dualconn.UdpHeader, payload []byte, err error) {

		if err != nil {
			if err != dualconn.ErrNotDestPort && err != dualconn.ErrSomethingIsWrong && err != dualconn.ErrBufsizeTooShort {
				fmt.Println(err)
			}
			return
		}

		addr, err := dualconn.NewSrcAddr(iph, udph)
		if err != nil {
			fmt.Println(err)
			return
		}

		spayload := string(payload)
		if strings.HasPrefix(spayload, "say") == false {
			fmt.Println("...... ", addr.String())
			return
		}

		sayHello(addr)
	}

	dcnn, err := dualconn.NewDualConn(ctx, *Port, *Bufsize, handler)

	if err != nil {
		log.Fatalln(err)
	}
	defer dcnn.Close()

	fmt.Println("enter q    -> os.exit")
	fmt.Println("enter c    -> cancel")

	var a string

	go func() {
		for {
			fmt.Scan(&a)
			if a == "q" {
				fmt.Println("exit....")
				os.Exit(0)
				continue
			}
			if a == "c" {
				fmt.Println("cancel....")
				cancel()
				continue
			}
		}
	}()

	for {
		select {
		case <-ctx.Done():
			fmt.Println("bye")
			return
		}
	}

}

func sayHello(addr net.Addr) {
	conn, err := net.Dial("udp", addr.String())
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	msg := "hello"

	conn.Write([]byte(msg))

}
