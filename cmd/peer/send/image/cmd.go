package image

import (
	"encoding/json"
	"image"
	"net/http"
	"os"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/ArminGh02/golang-p2p-messenger/internal/imgutil"
	"github.com/ArminGh02/golang-p2p-messenger/internal/protocol"
	"github.com/ArminGh02/golang-p2p-messenger/internal/response"
)

func NewCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "image <target username> <image filename>",
		Short: "send the specified image file to the specified username in a P2P way",
		RunE:  run,
		Args:  cobra.MatchAll(cobra.ExactArgs(2), validateArgs),
	}
}

func validateArgs(cmd *cobra.Command, args []string) error {
	if !isValidUsername(args[0]) || !isValidImage(args[1]) {
		return errors.New("invalid username or image filename")
	}
	return nil
}

func isValidUsername(u string) bool {
	return true
}

func isValidImage(i string) bool {
	return true
}

func run(cmd *cobra.Command, args []string) error {
	stunAddr, err := cmd.Flags().GetString("server")
	if err != nil {
		panic(err)
	}

	username, err := cmd.Flags().GetString("username")
	if err != nil {
		panic(err)
	}

	var (
		targetUsername = args[0]
		imageFilename  = args[1]
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

	f, err := os.Open(imageFilename)
	if err != nil {
		cmd.Println("could not open the file")
		return err
	}

	cmd.Println("opened the file")

	targetAddr := respBody.Peers[0].UDPAddr

	img, _, err := image.Decode(f)
	if err != nil {
		return err
	}

	pixels := imgutil.ToPixels(img)

	cmd.Println("sending...")
	return protocol.SendImage(targetAddr, pixels, imageFilename, username)
}
