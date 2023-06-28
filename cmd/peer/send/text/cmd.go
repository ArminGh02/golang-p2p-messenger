package text

import (
	"encoding/json"
	"net/http"

	"github.com/spf13/cobra"
	"github.com/pkg/errors"

	"github.com/ArminGh02/golang-p2p-messenger/internal/protocol"
	"github.com/ArminGh02/golang-p2p-messenger/internal/response"
)

func NewCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "text <target username> <desired text>",
		Short: "send a text to specified username in a P2P way",
		RunE:  run,
		Args:  cobra.ExactArgs(2),
	}
}

func run(cmd *cobra.Command, args []string) error {
	stunAddr, err := cmd.Flags().GetString("server")
	if err != nil {
		panic(err)
	}

	var (
		targetUsername = args[0]
		text           = args[1]
	)

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

	if resp.StatusCode != 200 {
		return errors.Errorf("failed to get peer from server at %s with status: %s", stunAddr, resp.Status)
	}

	var respBody response.GetPeer
	if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
		return err
	}

	if !respBody.OK {
		return errors.Errorf(
			"something went wrong retrieving peer with username %s on server at %s. error: %s",
			targetUsername,
			stunAddr,
			respBody.Error,
		)
	}

	return protocol.SendText(respBody.Peers[0].TCPAddr, text)
}
