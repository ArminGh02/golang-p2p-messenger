package peer

type Peer struct {
	UDPAddr  string `json:"udp_addr"`
	TCPAddr  string `json:"tcp_addr"`
	Username string `json:"username"`
}
