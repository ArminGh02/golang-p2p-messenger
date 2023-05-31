package list

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
		Use:   "list",
		Short: "lists all peers in the network",
		RunE:  run,
	}
}

func run(cmd *cobra.Command, args []string) error {
	stunURL, err := cmd.PersistentFlags().GetString("server")
	if err != nil {
		panic(err)
	}

	resp, err := requester.GetPeer(stunURL, "")
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
		return errors.Errorf("something went wrong when trying to get the list of peers from the server at %s. error: %s", stunURL, respBody.Error)
	}

	cmd.Printf("peers list: %#v\n", respBody.Peers)
	return nil
}
