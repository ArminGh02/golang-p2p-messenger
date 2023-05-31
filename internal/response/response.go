package response

import "github.com/ArminGh02/golang-p2p-messenger/internal/peer"

type (
	PostPeer struct {
		OK    bool   `json:"ok"`
		Error string `json:"error,omitempty"`
	}
	GetPeer struct {
		OK    bool         `json:"ok"`
		Error string       `json:"error,omitempty"`
		Peers []*peer.Peer `json:"peers,omitempty"`
	}
)
