package render

import (
	"context"
	"testing"
	"video-terminal/types"
)

func TestScanlineEffectFastIsNoop(t *testing.T) {
	in := types.WorkRGB{W: 2, H: 2, Stride: 6, Pix: []byte{100, 100, 100, 120, 120, 120, 140, 140, 140, 160, 160, 160}}
	out, err := ScanlineEffect{}.Apply(context.Background(), in, types.PresetFast)
	if err != nil {
		t.Fatalf("Apply returned error: %v", err)
	}
	for i := range in.Pix {
		if out.Pix[i] != in.Pix[i] {
			t.Fatalf("expected noop for fast preset at byte %d: got %d want %d", i, out.Pix[i], in.Pix[i])
		}
	}
}

func TestScanlineEffectCRTDimsAlternateRows(t *testing.T) {
	in := types.WorkRGB{W: 2, H: 2, Stride: 6, Pix: []byte{200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200}}
	out, err := ScanlineEffect{}.Apply(context.Background(), in, types.PresetCRT)
	if err != nil {
		t.Fatalf("Apply returned error: %v", err)
	}
	if out.Pix[0] <= out.Pix[6] {
		t.Fatalf("expected odd row to be dimmer than even row, got %d <= %d", out.Pix[0], out.Pix[6])
	}
	if out.Pix[0] == out.Pix[3] && out.Pix[1] == out.Pix[4] && out.Pix[2] == out.Pix[5] {
		t.Fatalf("expected CRT phosphor mask to alter adjacent columns")
	}
}

func BenchmarkScanlineCRT(b *testing.B) {
	in := types.WorkRGB{W: 160, H: 80, Stride: 480, Pix: make([]byte, 160*80*3)}
	for i := range in.Pix {
		in.Pix[i] = byte(i)
	}

	e := ScanlineEffect{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := e.Apply(context.Background(), in, types.PresetCRT); err != nil {
			b.Fatal(err)
		}
	}
}
