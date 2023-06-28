package root

import (
	"context"
	"fmt"
	"net"

	"github.com/ArminGh02/golang-p2p-messenger/internal/protocol"
)

func loopReceiveText(ctx context.Context, out chan<- string) error {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", tcpPort))
	if err != nil {
		return err
	}

	defer listener.Close()

	type connErrPair struct {
		conn net.Conn
		err  error
	}
	conns := make(chan connErrPair)
	go func(conns chan<- connErrPair) {
		for {
			conn, err := listener.Accept()
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
