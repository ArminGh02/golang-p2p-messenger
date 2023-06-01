package root

import (
	"context"
	"fmt"
	"image"
	"io"
	"net"
	"os"

	homedir "github.com/mitchellh/go-homedir"
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
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := homedir.Dir()
		cobra.CheckErr(err)

		viper.AddConfigPath(home)
		viper.AddConfigPath(".")
		viper.SetConfigName("config")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}

func NewCommand() *cobra.Command {
	cobra.OnInitialize(initConfig)

	cmd := &cobra.Command{
		Use:   "messenger",
		Short: "Command line p2p messenger",
		RunE:  run,
	}

	cmd.Flags().Uint16VarP(&udpPort, "tcp-port", "t", 8081, "TCP port to listen on")
	cmd.Flags().Uint16VarP(&tcpPort, "udp-port", "u", 8082, "UDP port to listen on")
	cmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (default is $HOME/config.yaml and current directory)")

	viper.BindPFlag("tcp-port", cmd.Flags().Lookup("tcp-port"))
	viper.BindPFlag("udp-port", cmd.Flags().Lookup("udp-port"))

	logger = logrus.New()
	logger.Out = cmd.OutOrStdout()

	return cmd
}

func run(cmd *cobra.Command, args []string) error {
	txtChan := make(chan string)
	imgChan := make(chan image.Image)

	defer close(txtChan)
	defer close(imgChan)

	go loopPrintOutput(cmd.Context(), cmd.OutOrStderr(), txtChan, imgChan)
	go loopRunCommand(cmd.Context())

	group, ctx := errgroup.WithContext(cmd.Context())
	group.Go(func() error { return loopReceiveText(ctx, txtChan) })
	group.Go(func() error { return loopReceiveImage(ctx, imgChan) })

	return group.Wait()
}

func loopRunCommand(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			if err := peer.NewCommand().Execute(); err != nil {
				logger.Errorln("Error executing command:", "error", err)
			}
		}
	}
}

func loopReceiveText(ctx context.Context, out chan<- string) error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", tcpPort))
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
		case <-imgChan:
			// store image and print path
		case txt := <-txtChan:
			fmt.Fprintf(out, "received message: %q", txt)
		}
	}
}
