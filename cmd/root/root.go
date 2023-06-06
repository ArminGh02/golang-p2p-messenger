package root

import (
	"bufio"
	"context"
	"fmt"
	"image"
	"net"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/sync/errgroup"

	"github.com/ArminGh02/golang-p2p-messenger/cmd/peer"
	"github.com/ArminGh02/golang-p2p-messenger/internal/protocol"
)

var (
	logger *logrus.Logger

	tcpPort uint16
	udpPort uint16
	cfgFile string
)

func initConfig() {
	if cfgFile == "" {
		cfgFile = "config.yaml"
	}

	viper.SetConfigFile(cfgFile)
	viper.AutomaticEnv()
	viper.ReadInConfig()
}

func NewCommand(exitCmd *cobra.Command) *cobra.Command {
	cobra.OnInitialize(initConfig)

	cmd := &cobra.Command{
		Use:   "messenger",
		Short: "Command line p2p messenger",
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd, args, exitCmd)
		},
	}

	cmd.Flags().Uint16VarP(&tcpPort, "tcp-port", "t", 8081, "TCP port to listen on")
	cmd.Flags().Uint16VarP(&udpPort, "udp-port", "u", 8082, "UDP port to listen on")
	cmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (default is $HOME/config.yaml and current directory)")

	viper.BindPFlag("tcp-port", cmd.Flags().Lookup("tcp-port"))
	viper.BindPFlag("udp-port", cmd.Flags().Lookup("udp-port"))

	logger = logrus.New()
	logger.Out = cmd.OutOrStdout()

	return cmd
}

func run(cmd *cobra.Command, args []string, exitCmd *cobra.Command) error {
	txtChan := make(chan string)
	imgChan := make(chan image.Image)

	defer close(txtChan)
	defer close(imgChan)

	go loopPrintOutput(cmd, txtChan, imgChan)
	go loopRunCommand(cmd, exitCmd)

	group, ctx := errgroup.WithContext(cmd.Context())
	group.Go(func() error { return loopReceiveText(ctx, txtChan) })
	group.Go(func() error { return loopReceiveImage(ctx, imgChan) })

	return group.Wait()
}

func loopRunCommand(cmd *cobra.Command, exitCmd *cobra.Command) {
	lines := make(chan string)

	go func(lines chan<- string) {
		s := bufio.NewScanner(os.Stdin)
		for s.Scan() {
			lines <- s.Text()
		}
	}(lines)

	for {
		cmd.Print(prompt() + " ")
		select {
		case <-cmd.Context().Done():
			return
		case line := <-lines:
			peerCmd := peer.NewCommand(exitCmd)
			args := strings.Fields(line)
			peerCmd.SetArgs(args)
			if err := peerCmd.Execute(); err != nil {
				logger.Errorln("Error executing command:", "error", err)
				// break
			}
			// if err := viper.WriteConfigAs("config.yaml"); err != nil {
			// 	logger.Errorln("Error writing config", "error", err)
			// }
		}
	}
}

func loopReceiveText(ctx context.Context, out chan<- string) error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", tcpPort))
	if err != nil {
		return err
	}

	type connErrPair struct {
		conn net.Conn
		err  error
	}
	conns := make(chan connErrPair)
	go func(conns chan<- connErrPair) {
		for {
			conn, err := lis.Accept()
			conns <- connErrPair{conn, err}
		}
	}(conns)

	errs := make(chan error)

	for {
		select {
		case <-ctx.Done():
			return nil
		case connErr := <-conns:
			conn, err := connErr.conn, connErr.err
			if err != nil {
				return err
			}

			go func() {
				defer conn.Close()
				txt, err := protocol.ReceiveText(conn)
				if err != nil {
					errs <- err
					return
				}
				out <- string(txt)
			}()
		case err := <-errs:
			return err
		}
	}
}

func loopReceiveImage(ctx context.Context, out chan<- image.Image) error {
	return nil
}

func loopPrintOutput(cmd *cobra.Command, txtChan <-chan string, imgChan <-chan image.Image) {
	for {
		select {
		case <-cmd.Context().Done():
			return
		case <-imgChan:
			// store image and print path
		case txt := <-txtChan:
			cmd.Printf("\rreceived message: %q\n%s ", txt, prompt())
		}
	}
}
