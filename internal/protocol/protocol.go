package protocol

import (
	"bytes"
	"encoding/binary"
	"image"
	"net"
	"os"
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

	addr, err := net.ResolveUDPAddr("udp", targetAddr)
	if err != nil {
		return err
	}

	conn, err := net.DialUDP("udp", nil, addr)
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
	addr, err := net.ResolveTCPAddr("tcp", targetAddr)
	if err != nil {
		return err
	}

	conn, err := net.DialTCP("tcp", nil, addr)
	if err != nil {
		return err
	}

	defer conn.Close()

	var buf bytes.Buffer
	binary.Write(&buf, binary.BigEndian, uint64(len(text)))
	buf.Write([]byte(text))
	_, err = conn.Write(buf.Bytes())
	return err
}

func ReceiveText(conn net.Conn) (res []byte, err error) {
	var buf [64]byte
	n, err := conn.Read(buf[:])
	if err != nil || n != 64 {
		return
	}

	msgLen := binary.BigEndian.Uint64(buf[:])
	res = make([]byte, msgLen)
	n, err = conn.Read(res)
	if err != nil || uint64(n) != msgLen {
		return nil, err
	}

	return
}
