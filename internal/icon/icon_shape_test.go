package icon

import (
	"bytes"
	"image/color"
	"image/png"
	"runtime"
	"testing"
)

func TestRoundedSquareShape(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shape test requires PNG decoding")
	}

	data, err := dotPNG(mustHexColor(colorHigh))
	if err != nil {
		t.Fatalf("render icon: %v", err)
	}

	img, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("decode png: %v", err)
	}

	center := toNRGBA(img.At(16, 16))
	if center.A != 0xff {
		t.Fatalf("expected center pixel to be opaque, got alpha %d", center.A)
	}

	edge := toNRGBA(img.At(6, 16))
	if edge.A != 0xff {
		t.Fatalf("expected edge pixel to be opaque, got alpha %d", edge.A)
	}

	corner := toNRGBA(img.At(6, 6))
	if corner.A != 0 {
		t.Fatalf("expected corner pixel to be transparent, got alpha %d", corner.A)
	}
}

func toNRGBA(c color.Color) color.NRGBA {
	return color.NRGBAModel.Convert(c).(color.NRGBA)
}
