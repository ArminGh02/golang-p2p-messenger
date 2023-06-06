package exit

import (
	"context"

	"github.com/spf13/cobra"
)

func NewCommand(cancel context.CancelFunc) *cobra.Command {
	return &cobra.Command{
		Use:     "exit",
		Short:   "exit",
		Aliases: []string{"q", "quit", "end"},
		Run: func(cmd *cobra.Command, args []string) {
			cancel()
		},
	}
}
