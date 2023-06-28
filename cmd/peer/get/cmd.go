package get

import (
	"encoding/json"
	"net/http"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/ArminGh02/golang-p2p-messenger/internal/response"
)

var (
	logger *logrus.Logger

	all bool
)

func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <target username> OR get --all",
		Short: "get the peer with specified username",
		RunE:  run,
		Args:  validateArgs,
	}

	cmd.Flags().BoolVarP(&all, "all", "a", false, "list all peers")

	logger = logrus.New()
	logger.Out = cmd.OutOrStdout()

	return cmd
}

func validateArgs(cmd *cobra.Command, args []string) error {
	if len(args) == 1 && all {
		return errors.New("a username argument is provided and --all flag is set simultaneously")
	}
	if len(args) == 0 && !all {
		return errors.New("--all flag is not set and no arguments for target username is provided")
	}
	return nil
}

func run(cmd *cobra.Command, args []string) error {
	stunAddr, err := cmd.Flags().GetString("server")
	if err != nil {
		panic(err)
	}

	var targetUsername string
	if !all {
		targetUsername = args[0]
	}

	resp, err := http.Get(stunAddr + "/peer/" + targetUsername)
	if err != nil {
		return errors.Wrapf(
			err,
			"failed to make the get request for target username %q from server at %q",
			targetUsername,
			stunAddr,
		)
	}

	defer resp.Body.Close()

	var respBody response.GetPeer
	if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
		return err
	}

	if resp.StatusCode == http.StatusNotFound {
		return errors.Errorf(
			"peer with username %s not found on the server at %s with status %s and error: %s",
			targetUsername,
			stunAddr,
			resp.Status,
			respBody.Error,
		)
	}

	if resp.StatusCode != http.StatusOK || !respBody.OK {
		return errors.Errorf(
			"failed to get username %s from server at %s with status %s and error: %s",
			targetUsername,
			stunAddr,
			resp.Status,
			respBody.Error,
		)
	}

	if all {
		cmd.Printf("peers info: %v\n", respBody.Peers)
	} else {
		cmd.Printf("peer info: %v\n", respBody.Peers[0])
	}
	return nil
}
