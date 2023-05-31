package image

import "github.com/spf13/cobra"

func NewCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "send image|text <target username> <image filename>|<desired text>",
		Short: "sends text/image to specified username in a P2P way",
		RunE:  run,
	}
}

func run(cmd *cobra.Command, args []string) error {
	return nil
}
