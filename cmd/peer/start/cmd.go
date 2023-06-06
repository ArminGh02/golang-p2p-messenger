package start

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/ArminGh02/golang-p2p-messenger/internal/request"
	"github.com/ArminGh02/golang-p2p-messenger/internal/response"
)

func NewCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "start",
		Short: "start peer and connect to STUN with specified username",
		RunE:  run,
	}
}

func run(cmd *cobra.Command, args []string) error {
	username, err := cmd.Flags().GetString("username")
	if err != nil {
		panic(err)
	}

	if username == "" {
		return errors.New("username is empty")
	}

	stunURL, err := cmd.Flags().GetString("server")
	if err != nil {
		panic(err)
	}

	req := request.PostPeer{
		Username: username,
		TCPAddr:  fmt.Sprintf("localhost:%d", viper.GetUint16("tcp-port")),
		UDPAddr:  fmt.Sprintf("localhost:%d", viper.GetUint16("udp-port")),
	}

	body, err := json.Marshal(&req)
	if err != nil {
		return err
	}

	resp, err := http.Post(stunURL+"/peer/", "application/json", bytes.NewBuffer(body))
	if err != nil {
		return errors.Wrapf(err, "failed to connect to STUN server: %s", stunURL)
	}

	if resp.StatusCode == http.StatusConflict {
		return errors.Errorf("username already exists")
	}

	if resp.StatusCode != http.StatusOK {
		return errors.Errorf("failed to connect to STUN server at %s with status: %s", stunURL, resp.Status)
	}

	var respBody response.PostPeer
	if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
		return errors.Wrap(err, "failed to decode response body")
	}

	if !respBody.OK {
		return errors.Errorf("failed to connect to STUN server at %s: %s", stunURL, respBody.Error)
	}

	return nil
}
