package send

import (
	"github.com/spf13/cobra"

	"github.com/ArminGh02/golang-p2p-messenger/cmd/peer/send/image"
	"github.com/ArminGh02/golang-p2p-messenger/cmd/peer/send/text"
)

func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "send image <target username> <image filename> OR send text <target username> <desired text>",
		Short: "send text/image to specified username in a P2P way",
		RunE:  run,
	}

	cmd.AddCommand(
		image.NewCommand(),
		text.NewCommand(),
	)

	return cmd
}

func run(cmd *cobra.Command, args []string) error {
	return nil
}
