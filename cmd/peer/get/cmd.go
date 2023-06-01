package get

import (
	"encoding/json"
	"net/http"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/ArminGh02/golang-p2p-messenger/internal/requester"
	"github.com/ArminGh02/golang-p2p-messenger/internal/response"
)

func NewCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "get <target username>",
		Short: "get the peer with specified username",
		RunE:  run,
		Args:  cobra.ExactArgs(1),
	}
}

func run(cmd *cobra.Command, args []string) error {
	stunURL, err := cmd.Flags().GetString("server")
	if err != nil {
		panic(err)
	}

	targetUsername := args[0]

	resp, err := requester.GetPeer(stunURL, targetUsername)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.Errorf("failed to get peer from server at %s with status: %s", stunURL, resp.Status)
	}

	var respBody response.GetPeer
	if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
		return err
	}

	if !respBody.OK {
		return errors.Errorf(
			"peer with username %s not found on the server at %s. error: %s",
			targetUsername,
			stunURL,
			respBody.Error,
		)
	}

	cmd.Printf("peer info: %#v\n", respBody.Peers[0])
	return nil
}
