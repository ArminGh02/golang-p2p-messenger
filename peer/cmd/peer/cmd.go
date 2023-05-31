package peer

import (
	"context"
	"fmt"
	"image"
	"io"
	"net"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"

	"github.com/ArminGh02/golang-p2p-messenger/internal/protocol"
	"github.com/ArminGh02/golang-p2p-messenger/peer/cmd/peer/get"
	"github.com/ArminGh02/golang-p2p-messenger/peer/cmd/peer/list"
	"github.com/ArminGh02/golang-p2p-messenger/peer/cmd/peer/send"
	"github.com/ArminGh02/golang-p2p-messenger/peer/cmd/peer/start"
)

var (
	logger *logrus.Logger

	tcpPort *uint16
	udpPort *uint16
)

var (
	CurrentServer *string
	Username      *string
)

func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "peer --tcp-port <tcp-port> --udp-port <udp-port> --server <server> --username <username>",
		Short:   "Command line p2p messenger",
		PreRunE: run,
	}

	tcpPort = cmd.Flags().Uint16P("tcp-port", "t", 8081, "TCP port to listen on")
	udpPort = cmd.Flags().Uint16P("udp-port", "u", 8082, "UDP port to listen on")

	cmd.PersistentFlags().StringP("server", "s", "http://localhost:8080", "server to connect to")
	cmd.PersistentFlags().StringP("username", "n", "", "username to use")
	cmd.MarkPersistentFlagRequired("username")

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

func run(cmd *cobra.Command, args []string) error {
	if *Username == "" {
		return fmt.Errorf("username is required")
	}

	txtChan := make(chan string)
	imgChan := make(chan image.Image)

	defer close(txtChan)
	defer close(imgChan)

	go loopPrintOutput(cmd.Context(), cmd.OutOrStderr(), txtChan, imgChan)

	group, ctx := errgroup.WithContext(cmd.Context())
	group.Go(func() error { return loopReceiveText(ctx, txtChan) })
	group.Go(func() error { return loopReceiveImage(ctx, imgChan) })

	return group.Wait()
}

func loopReceiveText(ctx context.Context, out chan<- string) error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *tcpPort))
	if err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			conn, err := lis.Accept()
			if err != nil {
				return err
			}

			group, _ := errgroup.WithContext(ctx)
			group.Go(func() error {
				defer conn.Close()
				txt, err := protocol.ReceiveText(conn)
				if err != nil {
					return err
				}
				out <- string(txt)
				return nil
			})
			return group.Wait()
		}
	}
}

func loopReceiveImage(ctx context.Context, out chan<- image.Image) error {
	return nil
}

func loopPrintOutput(ctx context.Context, out io.Writer, txtChan <-chan string, imgChan <-chan image.Image) {
	for {
		select {
		case <-ctx.Done():
			return
		case _ = <-imgChan:
			// store image and print path
		case txt := <-txtChan:
			fmt.Fprintf(out, "received message: %q", txt)
		}
	}
}
