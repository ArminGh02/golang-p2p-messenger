package main

import (
	"context"
	"os"

	"github.com/sirupsen/logrus"

	"github.com/ArminGh02/golang-p2p-messenger/cmd/peer/exit"
	"github.com/ArminGh02/golang-p2p-messenger/cmd/root"
)

func main() {
	logger := logrus.New()
	logger.Out = os.Stdout

	ctx, cancel := context.WithCancel(context.Background())
	exitCmd := exit.NewCommand(cancel)
	if err := root.NewCommand(exitCmd).ExecuteContext(ctx); err != nil {
		logger.Fatalln("Error executing command:", "error", err)
	}
}
