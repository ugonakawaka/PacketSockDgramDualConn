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
	// Bufsize buffer size
	Bufsize = flag.Int("bsize", 1500, "recv buffer size")
	b       = flag.Bool("b", false, "payload trim right side LF")
)

func main() {
	flag.Parse()
	fmt.Println(*AppName, *Port, *Dev, *Bufsize, *b)

	ctx, cancel := context.WithCancel(context.Background())

	// udp handler
	handler := func(n int, iph *dualconn.IpHeader, udph *dualconn.UdpHeader, payload []byte, err error) {

		if err != nil {

			if err != dualconn.ErrNotDestPort && err != dualconn.ErrSomethingIsWrong {
				fmt.Println(err)
			}
			return
		}

		// プロトコルバージョンのチェック
		if iph.Ver == 4 {
			fmt.Println("version:4")
		} else if iph.Ver == 6 {
			fmt.Println("version:6")
		}

		fmt.Printf("ip header:[%v]\n", iph)
		fmt.Printf("udp header:[%v]\n", udph)
		fmt.Printf("payload size:[%v] body:[%v]\n", len(payload), string(payload))
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
