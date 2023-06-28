package protocol

import (
	"encoding/json"
	"fmt"
	"image/color"
	"log"
	"net"
	"runtime"
	"strconv"
	"time"

	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
)

const (
	UsernameMaxLength  = 64
	FilenameMaxLength  = 64
	PayloadPixelsCount = 256

	DefaultTimeout = 5 * time.Second
	ImageTimeout   = 30 * time.Second
)

type ImagePacket struct {
	Sender   [UsernameMaxLength]byte
	Filename [FilenameMaxLength]byte
	Width    uint64
	Height   uint64
	Row      uint64
	Offset   uint64
	Pixels   [PayloadPixelsCount]uint32
}

type ImageACKPacket struct {
	Username  [UsernameMaxLength]byte
	Filename  [FilenameMaxLength]byte
	Flag      bool
	AckNumber uint64
}

func SendImage(targetAddr string, pixels [][]color.RGBA, filename string, sender string) error {
	conn, err := net.DialTimeout("udp", targetAddr, DefaultTimeout)
	if err != nil {
		if err, ok := err.(net.Error); ok && err.Timeout() {
			return errors.Wrapf(err, "%s timeout reached when dialling %s", DefaultTimeout, targetAddr)
		}
		return err
	}

	defer conn.Close()

	var bFilename [FilenameMaxLength]byte
	copy(bFilename[:], filename)

	var bSender [UsernameMaxLength]byte
	copy(bSender[:], sender)

	var packetPixels [PayloadPixelsCount]uint32
	var packetsCount int

	var group errgroup.Group
	group.SetLimit(runtime.NumCPU())

	limiter := time.Tick(100 * time.Millisecond)

	for i, row := range pixels {
		for j, pix := range row {
			packetPixels[j%PayloadPixelsCount] = uint32(pix.R)<<24 + uint32(pix.G)<<16 + uint32(pix.B)<<8 + uint32(pix.A)

			if j%PayloadPixelsCount != PayloadPixelsCount-1 {
				if j != len(row)-1 {
					continue
				}
			}

			p := ImagePacket{
				Sender:   bSender,
				Filename: bFilename,
				Width:    uint64(len(pixels[0])),
				Height:   uint64(len(pixels)),
				Row:      uint64(i),
				Offset:   uint64(j / PayloadPixelsCount),
				Pixels:   packetPixels,
			}

			b, err := json.Marshal(p)
			if err != nil {
				panic(err)
			}

			packetsCount++

			group.Go(func() error {
				<-limiter

				conn.SetWriteDeadline(time.Now().Add(DefaultTimeout))
				_, err := conn.Write(b)

				// wait for ack receipt
				// if no ack received resend and retry for 2 times
				// if still no ack, cancel the whole operation

				return err
			})
		}
	}

	log.Printf("a total of %d packets sent\n", packetsCount)

	return group.Wait()
}

func SendText(targetAddr string, text string) error {
	conn, err := net.DialTimeout("tcp", targetAddr, DefaultTimeout)
	if err != nil {
		if err, ok := err.(net.Error); ok && err.Timeout() {
			return errors.Wrapf(err, "%s timeout reached when dialling %s", DefaultTimeout, targetAddr)
		}
		return err
	}

	defer conn.Close()

	msg := fmt.Sprintf("%064d%s", len(text), text)
	conn.SetWriteDeadline(time.Now().Add(DefaultTimeout))
	_, err = conn.Write([]byte(msg))
	if err != nil {
		if err, ok := err.(net.Error); ok && err.Timeout() {
			return errors.Wrapf(err, "%s timeout reached when writing message to %s", DefaultTimeout, targetAddr)
		}
	}
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
