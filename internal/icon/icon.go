package icon

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"runtime"
	"strconv"
)

const (
	colorEmpty   = "#2c303b"
	colorLow     = "#216e39"
	colorMid     = "#30a14e"
	colorHigh    = "#40c463"
	colorMax     = "#9be9a8"
	colorError   = "#ff0000"
	iconSize     = 32
	circleRadius = 11.0
)

func ForCount(count int) ([]byte, error) {
	return dotPNG(colorForCount(count))
}

func ForError() ([]byte, error) {
	return dotPNG(mustHexColor(colorError))
}

func colorForCount(count int) color.NRGBA {
	switch {
	case count <= 0:
		return mustHexColor(colorEmpty)
	case count >= 1 && count <= 3:
		return mustHexColor(colorLow)
	case count >= 4 && count <= 6:
		return mustHexColor(colorMid)
	case count >= 7 && count <= 9:
		return mustHexColor(colorHigh)
	default:
		return mustHexColor(colorMax)
	}
}

func dotPNG(dot color.NRGBA) ([]byte, error) {
	img := image.NewNRGBA(image.Rect(0, 0, iconSize, iconSize))
	center := float64(iconSize-1) / 2
	r2 := circleRadius * circleRadius

	for y := 0; y < iconSize; y++ {
		dy := float64(y) - center
		for x := 0; x < iconSize; x++ {
			dx := float64(x) - center
			if dx*dx+dy*dy <= r2 {
				img.SetNRGBA(x, y, dot)
			}
		}
	}

	var buf bytes.Buffer
	if err := encodeIcon(&buf, img); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func encodeIcon(buf *bytes.Buffer, img image.Image) error {
	if runtime.GOOS == "windows" {
		return encodeICO(buf, img)
	}
	return png.Encode(buf, img)
}

func encodeICO(buf *bytes.Buffer, img image.Image) error {
	var pngBuf bytes.Buffer
	if err := png.Encode(&pngBuf, img); err != nil {
		return err
	}
	pngData := pngBuf.Bytes()

	width := img.Bounds().Dx()
	height := img.Bounds().Dy()
	if width <= 0 || height <= 0 || width > 256 || height > 256 {
		return fmt.Errorf("invalid icon size %dx%d", width, height)
	}

	w := byte(width)
	h := byte(height)
	if width == 256 {
		w = 0
	}
	if height == 256 {
		h = 0
	}

	if err := writeUint16(buf, 0); err != nil {
		return err
	}
	if err := writeUint16(buf, 1); err != nil {
		return err
	}
	if err := writeUint16(buf, 1); err != nil {
		return err
	}
	if err := buf.WriteByte(w); err != nil {
		return err
	}
	if err := buf.WriteByte(h); err != nil {
		return err
	}
	if err := buf.WriteByte(0); err != nil {
		return err
	}
	if err := buf.WriteByte(0); err != nil {
		return err
	}
	if err := writeUint16(buf, 1); err != nil {
		return err
	}
	if err := writeUint16(buf, 32); err != nil {
		return err
	}
	if err := writeUint32(buf, uint32(len(pngData))); err != nil {
		return err
	}
	if err := writeUint32(buf, 6+16); err != nil {
		return err
	}
	if _, err := buf.Write(pngData); err != nil {
		return err
	}
	return nil
}

func writeUint16(buf *bytes.Buffer, v uint16) error {
	_, err := buf.Write([]byte{byte(v), byte(v >> 8)})
	return err
}

func writeUint32(buf *bytes.Buffer, v uint32) error {
	_, err := buf.Write([]byte{
		byte(v),
		byte(v >> 8),
		byte(v >> 16),
		byte(v >> 24),
	})
	return err
}

func mustHexColor(hex string) color.NRGBA {
	if len(hex) != 7 || hex[0] != '#' {
		panic(fmt.Sprintf("invalid color: %s", hex))
	}

	r, err := strconv.ParseUint(hex[1:3], 16, 8)
	if err != nil {
		panic(err)
	}
	g, err := strconv.ParseUint(hex[3:5], 16, 8)
	if err != nil {
		panic(err)
	}
	b, err := strconv.ParseUint(hex[5:7], 16, 8)
	if err != nil {
		panic(err)
	}

	return color.NRGBA{R: uint8(r), G: uint8(g), B: uint8(b), A: 0xff}
}
