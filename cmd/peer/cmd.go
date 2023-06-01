package peer

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/ArminGh02/golang-p2p-messenger/cmd/peer/get"
	"github.com/ArminGh02/golang-p2p-messenger/cmd/peer/list"
	"github.com/ArminGh02/golang-p2p-messenger/cmd/peer/send"
	"github.com/ArminGh02/golang-p2p-messenger/cmd/peer/start"
)

var logger *logrus.Logger

func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "peer <command> [options]",
		Short: "Command line p2p messenger",
	}

	cmd.PersistentFlags().StringP("username", "n", "", "username to use")
	cmd.PersistentFlags().StringP("server", "s", "http://localhost:8080", "server to connect to")

	viper.BindPFlag("username", cmd.PersistentFlags().Lookup("username"))
	viper.BindPFlag("server", cmd.PersistentFlags().Lookup("server"))

	cmd.AddCommand(
		start.NewCommand(), // start connection to stun
		list.NewCommand(),  // list all peers
		get.NewCommand(),   // get peer by username
		send.NewCommand(),  // send image/text to a peer
	)

	logger = logrus.New()
	logger.Out = cmd.OutOrStdout()

	return cmd
}
