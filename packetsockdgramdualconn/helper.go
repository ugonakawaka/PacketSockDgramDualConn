package packetsockdgramdualconn

import (
	"net"
	"strconv"
)

// Addr net.Addr
type Addr struct {
	Version int
	Addr    string
	Port    int
}

// Network これは、net.Addrをみたす
func (a *Addr) Network() string {
	return "udp"
}

// String これは、net.Addrをみたす
func (a *Addr) String() string {

	if a.Version == 4 {
		return a.Addr + ":" + strconv.Itoa(a.Port)
	} else if a.Version == 6 {
		return "[" + a.Addr + "]:" + strconv.Itoa(a.Port)
	}

	return "error"
}

// NewSrcAddr helper function make net.Addr
func NewSrcAddr(iph *IpHeader, udph *UdpHeader) (naddr net.Addr, err error) {
	addr := new(Addr)
	addr.Version = iph.Ver
	if iph.Ver == Ipv4 {
		addr.Addr = iph.Ipv4Header.Src.String()
	} else if iph.Ver == Ipv6 {
		addr.Addr = iph.Ipv6Header.Src.String()
	} else {
		return nil, ErrSomethingIsWrong
	}

	addr.Port = udph.SourcePort

	return addr, nil
}
