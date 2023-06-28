package root

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/ArminGh02/golang-p2p-messenger/internal/imgutil"
	"image/color"
	"io"
	"log"
	"math"
	"net"

	"github.com/ArminGh02/golang-p2p-messenger/internal/protocol"
	"github.com/pkg/errors"
)

type packetData struct {
	row    uint64
	offset uint64
	pixels []color.RGBA
}

func loopReceiveImage(ctx context.Context, out chan<- imageData) error {
	conn, err := net.ListenPacket("udp", fmt.Sprintf(":%d", udpPort))
	logger.Infoln("listening udp")
	if err != nil {
		return errors.Wrapf(err, "unable to start listening on port %d for UDP packets", udpPort)
	}

	defer conn.Close()

	packetsChan := make(chan protocol.ImagePacket)
	go func() {
		type key struct {
			username [protocol.UsernameMaxLength]byte
			filename [protocol.FilenameMaxLength]byte
		}
		packets := make(map[key][]packetData)
		for {
			select {
			case imgPacket := <-packetsChan:
				key := key{imgPacket.Sender, imgPacket.Filename}

				allPacketsCount := imgPacket.Height * uint64(math.Ceil(float64(imgPacket.Width)/protocol.PayloadPixelsCount))
				if packets[key] == nil {
					packets[key] = make([]packetData, 0, allPacketsCount)
				}

				packets[key] = append(packets[key], pcktData(imgPacket))

				// mark packet (row,seq) as received so that we don't store duplicates
				// send ack for this packet

				logger.Infoln("len =", len(packets[key]), "allPacketsCount =", allPacketsCount)

				if uint64(len(packets[key])) == allPacketsCount {
					reassembleImage(
						packets[key],
						string(imgPacket.Sender[:]),
						string(imgPacket.Filename[:]),
						imgPacket.Width,
						imgPacket.Height,
						out,
					)
				}
			}
		}
	}()

	for {
		buf := make([]byte, 256*256)

		logger.Infoln("reading")

		n, _, err := conn.ReadFrom(buf)
		if err != nil {
			if err != io.EOF {
				logger.Error("read error:", err)
			}
		}

		go func() {
			logger.Infoln("read it")

			var imgPacket protocol.ImagePacket
			if err := json.Unmarshal(buf[:n], &imgPacket); err != nil {
				logger.Error(err)
			}

			packetsChan <- imgPacket
		}()
	}
}

func pcktData(imgPacket protocol.ImagePacket) packetData {
	var pcktData packetData
	pcktData.offset = imgPacket.Offset
	pcktData.row = imgPacket.Row
	pcktData.pixels = make([]color.RGBA, protocol.PayloadPixelsCount)
	for i := 0; i < protocol.PayloadPixelsCount; i++ {
		rgba := imgPacket.Pixels[i]
		r := uint8((rgba >> 24) & 0xFF)
		g := uint8((rgba >> 16) & 0xFF)
		b := uint8((rgba >> 8) & 0xFF)
		a := uint8(rgba & 0xFF)
		pcktData.pixels[i] = color.RGBA{R: r, G: g, B: b, A: a}
	}
	return pcktData
}

func reassembleImage(
	packets []packetData,
	username string,
	filename string,
	width uint64,
	height uint64,
	out chan<- imageData,
) {
	log.Println("height =", height, "width =", width)

	pixels := make([][]color.RGBA, height)
	for i := range pixels {
		pixels[i] = make([]color.RGBA, width)
	}

	for _, packet := range packets {
		for i := 0; i < protocol.PayloadPixelsCount; i++ {
			idx := packet.offset*protocol.PayloadPixelsCount + uint64(i)
			if idx < width {
				pixels[packet.row][idx] = packet.pixels[i]
			}
		}
	}

	img := imgutil.FromPixels(pixels)
	out <- imageData{
		Image:    img,
		filename: filename,
		username: username,
	}
}
