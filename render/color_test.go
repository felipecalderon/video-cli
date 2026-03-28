package render

import (
	"context"
	"testing"
	"video-terminal/types"
)

func TestChannelQuantizerQuantizesPerChannel(t *testing.T) {
	in := types.WorkRGB{
		W:      2,
		H:      1,
		Stride: 6,
		Pix:    []byte{10, 128, 250, 33, 77, 199},
	}

	out, err := ChannelQuantizer{}.Quantize(nil, in, types.ColorTruecolor)
	if err != nil {
		t.Fatalf("Quantize returned error: %v", err)
	}

	if got, want := out.Pix[0], quantizeChannel(10, quantizeTruecolorLevels); got != want {
		t.Fatalf("red channel mismatch: got %d want %d", got, want)
	}
	if got, want := out.Pix[1], quantizeChannel(128, quantizeTruecolorLevels); got != want {
		t.Fatalf("green channel mismatch: got %d want %d", got, want)
	}
	if got, want := out.Pix[2], quantizeChannel(250, quantizeTruecolorLevels); got != want {
		t.Fatalf("blue channel mismatch: got %d want %d", got, want)
	}
}

func TestBayerDitherAppliesSpatialVariation(t *testing.T) {
	in := types.WorkRGB{
		W:      4,
		H:      4,
		Stride: 12,
		Pix:    make([]byte, 4*4*3),
	}
	for i := range in.Pix {
		in.Pix[i] = 120
	}

	quantized, err := ChannelQuantizer{}.Quantize(nil, in, types.ColorTruecolor)
	if err != nil {
		t.Fatalf("Quantize returned error: %v", err)
	}

	out, err := BayerDither{}.Dither(nil, quantized, types.PresetFast)
	if err != nil {
		t.Fatalf("Dither returned error: %v", err)
	}

	if out.Pix[0] == out.Pix[3] && out.Pix[1] == out.Pix[4] && out.Pix[2] == out.Pix[5] {
		t.Fatalf("expected spatial variation after dithering, got identical neighboring pixels")
	}
}

func TestTileContrastFlatIsLow(t *testing.T) {
	in := types.WorkRGB{W: 3, H: 3, Stride: 9, Pix: make([]byte, 3*3*3)}
	for i := range in.Pix {
		in.Pix[i] = 80
	}
	luma := buildLumaBuffer(in)

	if got := tileContrast(luma, in.W, in.H, 0, 0, 2); got != 0 {
		t.Fatalf("expected zero contrast for flat area, got %d", got)
	}
}

func TestTileContrastEdgeIsHigh(t *testing.T) {
	in := types.WorkRGB{W: 4, H: 4, Stride: 12, Pix: make([]byte, 4*4*3)}
	for y := 0; y < 4; y++ {
		for x := 0; x < 4; x++ {
			idx := y*in.Stride + x*3
			val := uint8(0)
			if x >= 2 {
				val = 255
			}
			in.Pix[idx+0] = val
			in.Pix[idx+1] = val
			in.Pix[idx+2] = val
		}
	}
	luma := buildLumaBuffer(in)

	if got := tileContrast(luma, in.W, in.H, 0, 0, 4); got < 40 {
		t.Fatalf("expected noticeable contrast near edge, got %d", got)
	}
}

func BenchmarkBayerDitherQuality(b *testing.B) {
	in := types.WorkRGB{W: 160, H: 80, Stride: 480, Pix: make([]byte, 160*80*3)}
	for i := range in.Pix {
		in.Pix[i] = byte(i)
	}

	d := BayerDither{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := d.Dither(context.Background(), in, types.PresetQuality); err != nil {
			b.Fatal(err)
		}
	}
}
