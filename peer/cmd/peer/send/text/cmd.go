package text

import (
	"encoding/json"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/ArminGh02/golang-p2p-messenger/internal/protocol"
	"github.com/ArminGh02/golang-p2p-messenger/internal/requester"
	"github.com/ArminGh02/golang-p2p-messenger/internal/response"
)

func NewCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "text <target username> <desired text>",
		Short: "sends a text to specified username in a P2P way",
		RunE:  run,
		Args:  cobra.ExactArgs(2),
	}
}

func run(cmd *cobra.Command, args []string) error {
	stunURL, err := cmd.PersistentFlags().GetString("server")
	if err != nil {
		panic(err)
	}

	var (
		targetUsername = args[0]
		text           = args[1]
	)

	resp, err := requester.GetPeer(stunURL, targetUsername)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return errors.Errorf("failed to get peer from server at %s with status: %s", stunURL, resp.Status)
	}

	var respBody response.GetPeer
	if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
		return err
	}

	if !respBody.OK {
		return errors.Errorf(
			"something went wrong retrieving peer with username %s on server at %s. error: %s",
			targetUsername,
			stunURL,
			respBody.Error,
		)
	}

	return protocol.SendText(respBody.Peers[0].TCPAddr, text)
}
