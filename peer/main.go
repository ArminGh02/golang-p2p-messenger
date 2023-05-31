package main

import (
	"os"

	"github.com/ArminGh02/golang-p2p-messenger/peer/cmd/peer"

	"github.com/sirupsen/logrus"
)

func main() {
	// define some commands to
	//   connect to stun,
	//   get all peers,
	//   get a peer by username,
	//   establish connection to a peer
	//     then:
	//     send image
	//     send text

	logger := logrus.New()
	logger.Out = os.Stdout

	if err := peer.NewCommand().Execute(); err != nil {
		logger.Fatalln("Error executing command:", "error", err)
	}
}
