package request

type (
	PostPeer struct {
		UDPAddr  string `json:"udp_addr"`
		TCPAddr  string `json:"tcp_addr"`
		Username string `json:"username"`
	}
)
