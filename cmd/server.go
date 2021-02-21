// +build linux
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	dualconn "github.com/ugonakawaka/PacketSockDgramDualConn/packetsockdgramdualconn"
)

var (
	// AppName application name
	AppName = flag.String("name", "TEST UDP SERVER", "application name")
	// Port port
	Port = flag.Int("port", 55501, "message receive port")
	// Dev target device
	Dev = flag.String("if", "lo", "interface")
	// Bufsize sss
	Bufsize = flag.Int("buf sieze", 1500, "recv buffer size")
	b       = flag.Bool("b", false, "payload trim right side LF")
)

func main() {
	flag.Parse()
	fmt.Println(*AppName, *Port, *Dev, *Bufsize, *b)

	ctx, cancel := context.WithCancel(context.Background())

	// error disp level
	errl := 0

	// udp handler
	handler := func(n int, iph *dualconn.IpHeader, udph *dualconn.UdpHeader, payload []byte, err error) {

		if err != nil {
			if errl == 1 {
				fmt.Println(err)
			}
			if err != dualconn.ErrNotDestPort {
				fmt.Println(err)
			}
			return
		}

		fmt.Printf("size:[%d]\n", n)
		fmt.Printf("ip header:[%v]\n", iph)
		fmt.Printf("udp header:[%v]\n", udph)
		fmt.Printf("payload:[%v]\n", string(payload))
		// fmt.Printf("%s\n", hex.Dump(b[:n]))
		fmt.Println("")
	}

	dcnn, err := dualconn.NewDualConn(ctx, *Port, *Bufsize, handler)

	if err != nil {
		log.Fatalln(err)
	}
	defer dcnn.Close()

	fmt.Println("enter q    -> os.exit")
	fmt.Println("enter c    -> cancel")
	fmt.Println("enter 0 -> error no print")
	fmt.Println("enter 1 -> error print ")
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
			if a == "0" {
				errl = 0
				continue
			}
			if a == "1" {
				errl = 1
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
