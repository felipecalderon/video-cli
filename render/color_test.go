package render

import (
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
