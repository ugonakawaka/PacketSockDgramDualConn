// create:2021/02/20
// update:2021/02/20
package packetsockdgram

import (
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"os"
	"syscall"
	"unsafe"

	"golang.org/x/net/ipv4"
	"golang.org/x/net/ipv6"
	// "golang.org/x/net/internal/socket"
)

// ===================
// このソースは三つのパートに分かれています

// ===================

// ===================
// helper

//  package err definition
var (
	ErrNotDestPort     = errors.New("not match dest port")
	ErrInvalidConn     = errors.New("invalid connection")
	ErrMissingAddress  = errors.New("missing address")
	ErrNilHeader       = errors.New("nil header")
	ErrBufsizeTooShort = errors.New("buffer size too short")
	ErrHeaderTooShort  = errors.New("header too short")
	ErrPayloadLen      = errors.New("payload size not correct")
)

func htons(host uint16) uint16 {
	return (host&0xff)<<8 | (host >> 8)
}

// ===================
// header
const (
	UDPHeaderLen = 8
)

type IpHeader struct {
	Ipv4Header *ipv4.Header
	Ipv6Header *ipv6.Header
}
type UdpHeader struct {
	SourcePort      int
	DestinationPort int
	Length          int
	Checksum        int
}

func (h *UdpHeader) String() string {
	if h == nil {
		return "<nil>"
	}
	return fmt.Sprintf("srcport=%d dstport=%d len=%d Checksum=%#x", h.SourcePort, h.DestinationPort, h.Length, h.Checksum)
}

// RFC 768
func ParseUDPHeader(b []byte) (*UdpHeader, error) {
	if b == nil {
		return nil, ErrNilHeader
	}
	if len(b) < UDPHeaderLen {
		return nil, ErrHeaderTooShort
	}
	h := new(UdpHeader)

	h.SourcePort = int(binary.BigEndian.Uint16(b[0:2]))
	h.DestinationPort = int(binary.BigEndian.Uint16(b[2:4]))
	h.Length = int(binary.BigEndian.Uint16(b[4:6]))
	h.Checksum = int(binary.BigEndian.Uint16(b[6:8]))
	return h, nil
}

// ===================
// packet
type handler struct {
	fd      int
	rcvport int
}

func (c *handler) ok() bool { return c != nil }

func (hdl *handler) readFrom(b []byte) (n int, h *IpHeader, uh *UdpHeader, p []byte, err error) {
	if !hdl.ok() {
		return -1, nil, nil, nil, ErrInvalidConn
	}

	n, _, err = syscall.Recvfrom(hdl.fd, b, syscall.MSG_TRUNC)

	if err != nil {
		return -1, nil, nil, nil, err
	}

	// recv data size check
	if len(b) < n {
		return n, nil, nil, nil, ErrBufsizeTooShort
	}

	if b[0] == 0x45 {
		return hdl.readFromIpv4(b)
	} else if b[0] == 0x60 {
		return hdl.readFromIpv6(b)
	}
	return -1, nil, nil, nil, err
}

func (hdl *handler) readFromIpv4(b []byte) (n int, iph *IpHeader, uh *UdpHeader, p []byte, err error) {

	// ipheader
	ipv4h, err := ipv4.ParseHeader(b)
	if err != nil {
		return n, nil, nil, nil, err
	}

	iph = &IpHeader{Ipv4Header: ipv4h}
	endudp4h := ipv4.HeaderLen + UDPHeaderLen

	// udp header
	uh, err = ParseUDPHeader(b[ipv4.HeaderLen:endudp4h])

	if err != nil {
		return n, iph, nil, nil, err
	}

	pb := b[endudp4h:]

	// port check
	if hdl.rcvport != uh.DestinationPort {
		return n, iph, uh, pb, ErrNotDestPort
	}

	return n, iph, uh, pb, nil
}
func (hdl *handler) readFromIpv6(b []byte) (n int, iph *IpHeader, uh *UdpHeader, p []byte, err error) {

	// ipheader
	ipv6h, err := ipv6.ParseHeader(b)
	if err != nil {
		return n, nil, nil, nil, err
	}

	iph = &IpHeader{Ipv6Header: ipv6h}

	startudp6h := ipv6.HeaderLen
	endudp6h := startudp6h + UDPHeaderLen

	// udp header
	uh, err = ParseUDPHeader(b[startudp6h:])
	if err != nil {
		return n, iph, nil, nil, err
	}

	pb := b[endudp6h:]
	// port check
	if hdl.rcvport != uh.DestinationPort {
		return n, iph, uh, pb, ErrNotDestPort
	}

	return n, iph, uh, pb, nil
}

// ===================
// endpoint

const (
	Ipv4 = 4
	Ipv6 = 6
)

type DualConn struct {
	// genericOpt
	conn4 Conn
	conn6 Conn
}

type Conn struct {
	// genericOpt
	fd int
	handler
}

//
func (dc *DualConn) Close() {
	dc.conn4.Close()
	dc.conn6.Close()
}

//
func (c *Conn) Close() error {
	if !c.handler.ok() {
		return ErrInvalidConn
	}
	return syscall.Close(c.fd)
}

//
func NewDualConn(port int,
	bufsize int,
	handler func(int, *IpHeader, *UdpHeader, []byte, error)) (*DualConn, error) {
	return NewDualConnIf(port, "", bufsize, handler)
}

//
func NewDualConnIf(port int, ifname string,
	bufsize int,
	handler func(int, *IpHeader, *UdpHeader, []byte, error)) (*DualConn, error) {
	ifindex := 0
	if ifname != "" {
		intf, err := net.InterfaceByName(ifname)
		if err != nil {
			return nil, err
		}
		ifindex = intf.Index
	}

	conn4, err := NewConn(Ipv4, port, int32(ifindex))
	if err != nil {
		return nil, err
	}
	conn6, err := NewConn(Ipv6, port, int32(ifindex))
	if err != nil {
		conn4.Close()
		return nil, err
	}

	dcnn := &DualConn{}
	dcnn.conn4 = *conn4
	dcnn.conn6 = *conn6

	goRoutineReadfrom(conn4, bufsize, handler)
	goRoutineReadfrom(conn6, bufsize, handler)

	return dcnn, nil
}

func goRoutineReadfrom(cnn *Conn, bufsize int, handler func(int, *IpHeader, *UdpHeader, []byte, error)) {
	go func() {
		for {
			b := make([]byte, bufsize)
			n, iph, udph, payload, err := cnn.Readfrom(b)
			handler(n, iph, udph, payload, err)
		}
	}()
}

//
func NewConn(protco int, port int, ifindex int32) (*Conn, error) {
	ethP := func(p int) int {

		switch p {
		case Ipv4: //
			return syscall.ETH_P_IP
		case Ipv6: //
			return syscall.ETH_P_IPV6
		default:
			return syscall.ETH_P_ALL
		}

	}(protco)

	fd, err := syscall.Socket(syscall.AF_PACKET, syscall.SOCK_DGRAM, 0)
	if err != nil {
		return nil, err
	}

	sll := syscall.RawSockaddrLinklayer{
		Family:   syscall.AF_PACKET,
		Protocol: htons(uint16(ethP)),
		// sll_ifindex 0 matches any interface (only permitted forbinding).
		Ifindex: ifindex,
	}

	_, _, err = syscall.Syscall(syscall.SYS_BIND, uintptr(fd),
		uintptr(unsafe.Pointer(&sll)), unsafe.Sizeof(sll))
	if errors.Is(err, os.ErrNotExist) {
		return nil, err
	}

	cnn := &Conn{
		fd:      fd,
		handler: handler{fd, port},
	}
	return cnn, nil
}

func (c *Conn) Readfrom(b []byte) (n int, h *IpHeader, uh *UdpHeader, p []byte, err error) {
	return c.handler.readFrom(b)
}

// ===================
