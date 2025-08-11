package root

import (
	"context"
	"encoding/json"
	"fmt"
	"image/color"
	"io"
	"log"
	"math"
	"net"

	"github.com/pkg/errors"

	"github.com/ArminGh02/golang-p2p-messenger/internal/imgutil"
	"github.com/ArminGh02/golang-p2p-messenger/internal/protocol"
)

type storedPacket struct {
	row    uint64
	offset uint64
	pixels []color.RGBA
}

type rowOffsetPair struct {
	row    uint64
	offset uint64
}

type userFilePair struct {
	username string
	filename string
}

func loopReceiveImage(ctx context.Context, out chan<- imageData) error {
	conn, err := net.ListenPacket("udp", fmt.Sprintf(":%d", udpPort))
	if err != nil {
		return errors.Wrapf(err, "unable to start listening on port %d for UDP packets", udpPort)
	}
	defer conn.Close()

	type sharedPacket struct {
		imgPacket protocol.ImagePacket
		addr      net.Addr
	}
	packetsChan := make(chan sharedPacket)
	defer close(packetsChan)

	go func() {
		packets := make(map[userFilePair]map[rowOffsetPair]storedPacket)
		for p := range packetsChan {
			imgPacket := p.imgPacket
			addr := p.addr

			key := userFilePair{imgPacket.Sender, imgPacket.Filename}
			rowOffset := rowOffsetPair{imgPacket.Row, imgPacket.Offset}

			allPacketsCount := imgPacket.Height * uint64(math.Ceil(float64(imgPacket.Width)/protocol.PayloadPixelsCount))
			if packets[key] == nil {
				packets[key] = make(map[rowOffsetPair]storedPacket, allPacketsCount)
			}

			if _, duplicate := packets[key][rowOffset]; duplicate {
				continue
			}

			packets[key][rowOffset] = toStoredPacket(imgPacket)

			ack(conn, addr, imgPacket)

			logger.Infof("packet %d from %d\n", len(packets[key]), allPacketsCount)

			if uint64(len(packets[key])) == allPacketsCount {
				reassembleImage(
					packets[key],
					imgPacket.Sender,
					imgPacket.Filename,
					imgPacket.Width,
					imgPacket.Height,
					out,
				)
			}
		}
	}()

	for {
		buf := make([]byte, 256*256)

		n, addr, err := conn.ReadFrom(buf)
		if err != nil {
			if err != io.EOF {
				logger.Error("read error:", err)
			}
		}

		go func() {
			var imgPacket protocol.ImagePacket
			if err := json.Unmarshal(buf[:n], &imgPacket); err != nil {
				logger.Error(err)
			}

			packetsChan <- sharedPacket{imgPacket, addr}
		}()
	}
}

func ack(conn net.PacketConn, addr net.Addr, imgPacket protocol.ImagePacket) {
	ack := protocol.ImageACKPacket{
		Username: imgPacket.Sender,
		Filename: imgPacket.Filename,
		Flag:     true,
		Offset:   imgPacket.Offset,
		Row:      imgPacket.Row,
	}

	b, err := json.Marshal(ack)
	if err != nil {
		panic(err)
	}

	_, err = conn.WriteTo(b, addr)
	if err != nil {
		logger.Error(err)
	}
}

func toStoredPacket(imgPacket protocol.ImagePacket) storedPacket {
	var pcktData storedPacket
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
	packets map[rowOffsetPair]storedPacket,
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
			col := packet.offset*protocol.PayloadPixelsCount + uint64(i)
			pixels[packet.row][col] = packet.pixels[i]
		}
	}

	img := imgutil.FromPixels(pixels)
	out <- imageData{
		Image:    img,
		filename: filename,
		username: username,
	}
}
