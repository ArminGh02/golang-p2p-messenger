package requester

import (
	"net/http"

	"github.com/pkg/errors"
)

func GetPeer(serverAddr, targetUsername string) (resp *http.Response, err error) {
	url := serverAddr + "/peer"
	if targetUsername != "" {
		url += "/" + targetUsername
	}

	resp, err = http.Get(url)
	if err != nil {
		return nil, errors.Wrapf(
			err,
			"failed to make the get request for target username %q from server at %q",
			targetUsername,
			serverAddr,
		)
	}

	return
}
