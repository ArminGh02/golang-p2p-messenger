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

	MaxRetries = 3
	opt = false
)

type ImagePacket struct {
	Sender   string
	Filename string
	Width    uint64
	Height   uint64
	Row      uint64
	Offset   uint64
	Pixels   [PayloadPixelsCount]uint32
}

type ImageACKPacket struct {
	Username string
	Filename string
	Flag     bool
	Row      uint64
	Offset   uint64
}

func SendImage(targetAddr string, pixels [][]color.RGBA, filename string, sender string) error {
	if len(filename) > FilenameMaxLength {
		return errors.New("filename length exceeded")
	}

	if len(sender) > UsernameMaxLength {
		return errors.New("username length exceeded")
	}

	conn, err := net.DialTimeout("udp", targetAddr, DefaultTimeout)
	if err != nil {
		if err, ok := err.(net.Error); ok && err.Timeout() {
			return errors.Wrapf(err, "%s timeout reached when dialling %s", DefaultTimeout, targetAddr)
		}
		return err
	}

	defer conn.Close()

	type rowOffsetPair struct {
		row    uint64
		offset uint64
	}
	acks := make(map[rowOffsetPair]bool)
	go func() {
		buf := make([]byte, 1024)

		conn.SetReadDeadline(time.Now().Add(DefaultTimeout))
		n, err := conn.Read(buf)
		if err != nil {
			// fix:
			panic(err)
		}

		var pckt ImageACKPacket
		if err := json.Unmarshal(buf[:n], &pckt); err != nil {
			panic(err)
		}

		acks[rowOffsetPair{pckt.Row, pckt.Offset}] = true
	}()

	var packetPixels [PayloadPixelsCount]uint32
	var packetsCount int

	var group errgroup.Group
	group.SetLimit(runtime.NumCPU())

	limiter := time.NewTicker(100 * time.Millisecond)
	errorChan := make(chan error, 1)

	go func() {
		for i, row := range pixels {
			for j, pix := range row {
				packetPixels[j%PayloadPixelsCount] = uint32(pix.R)<<24 + uint32(pix.G)<<16 + uint32(pix.B)<<8 + uint32(pix.A)

				if j%PayloadPixelsCount != PayloadPixelsCount-1 {
					if j != len(row)-1 {
						continue
					}
				}

				pckt := ImagePacket{
					Sender:   sender,
					Filename: filename,
					Width:    uint64(len(pixels[0])),
					Height:   uint64(len(pixels)),
					Row:      uint64(i),
					Offset:   uint64(j / PayloadPixelsCount),
					Pixels:   packetPixels,
				}

				b, err := json.Marshal(pckt)
				if err != nil {
					panic(err)
				}

				packetsCount++

				group.Go(func() error {
					<-limiter.C

					if opt {
						var ack bool
					RetryLoop:
						for i := 0; i < MaxRetries; i++ {
							conn.SetWriteDeadline(time.Now().Add(DefaultTimeout))
							_, err := conn.Write(b)
							if err != nil {
								return err
							}

							for i, t := 0, time.NewTicker(2*DefaultTimeout/20); i < 20; i++ {
								<-t.C
								if acks[rowOffsetPair{pckt.Row, pckt.Offset}] {
									ack = true
									break RetryLoop
								}
							}
						}
						if !ack {
							err := errors.New("3 retries and still no ack received")
							errorChan <- err
							return err
						}
					} else {
						conn.SetWriteDeadline(time.Now().Add(DefaultTimeout))
						_, err := conn.Write(b)
						if err != nil {
							return err
						}
					}
					return nil
				})
			}
		}

		log.Printf("a total of %d packets sent\n", packetsCount)

		errorChan <- group.Wait()
	}()

	for err := range errorChan {
		return err
	}

	return nil
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
