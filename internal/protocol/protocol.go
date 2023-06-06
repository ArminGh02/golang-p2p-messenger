package protocol

import (
	"bytes"
	"fmt"
	"image"
	"net"
	"os"
	"strconv"

	"github.com/pkg/errors"
)

func SendImage(targetAddr string, filename string) error {
	f, err := os.Open(filename)
	if err != nil {
		return err
	}

	defer f.Close()

	_, _, err = image.Decode(f)
	if err != nil {
		return err
	}

	conn, err := net.Dial("udp", targetAddr)
	if err != nil {
		return err
	}

	defer conn.Close()

	// checksum
	// row number
	// what to do when a packet is lost?
	return nil
}

func SendText(targetAddr string, text string) error {
	conn, err := net.Dial("tcp", targetAddr)
	if err != nil {
		return err
	}

	defer conn.Close()

	var buf bytes.Buffer
	buf.Write([]byte(fmt.Sprintf("%064d", len(text))))
	buf.Write([]byte(text))
	_, err = conn.Write(buf.Bytes())
	return err
}

func ReceiveText(conn net.Conn) (res []byte, err error) {
	buf := make([]byte, 64)
	n, err := conn.Read(buf)
	if err != nil {
		return
	}
	if n != 64 {
		err = errors.New("could not read 64 header bytes")
		return
	}

	msgLen, err := strconv.Atoi(string(buf))
	if err != nil {
		panic(err)
	}

	res = make([]byte, msgLen)
	n, err = conn.Read(res)
	if err != nil {
		res = nil
		return
	}
	if n != msgLen {
		return nil, errors.New("could not read the whole message")
	}

	return
}
