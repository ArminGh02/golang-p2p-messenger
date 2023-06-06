package peer

import "fmt"

type Peer struct {
	UDPAddr  string `json:"udp_addr"`
	TCPAddr  string `json:"tcp_addr"`
	Username string `json:"username"`
}

func (p *Peer) String() string {
	return fmt.Sprintf("Peer{Username:%s, UDPAddr:%s, TCPAddr:%s}", p.Username, p.UDPAddr, p.TCPAddr)
}
