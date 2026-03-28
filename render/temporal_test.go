package render

import (
	"context"
	"reflect"
	"testing"
	"video-terminal/types"
)

func TestTemporalBlendAlphaZero(t *testing.T) {
	b := &TemporalBlend{}
	in := workRGB(1, 1, []uint8{10, 20, 30})
	out, err := b.Blend(context.Background(), in, 0)
	if err != nil {
		t.Fatalf("blend error: %v", err)
	}
	if !reflect.DeepEqual(out.Pix, []uint8{10, 20, 30}) {
		t.Fatalf("unexpected output: %v", out.Pix)
	}
}

func TestTemporalBlendAlphaOneUsesPrev(t *testing.T) {
	b := &TemporalBlend{}
	in1 := workRGB(1, 1, []uint8{10, 20, 30})
	_, _ = b.Blend(context.Background(), in1, 0.5)

	in2 := workRGB(1, 1, []uint8{200, 210, 220})
	out, err := b.Blend(context.Background(), in2, 1)
	if err != nil {
		t.Fatalf("blend error: %v", err)
	}
	if !reflect.DeepEqual(out.Pix, []uint8{10, 20, 30}) {
		t.Fatalf("expected prev frame, got: %v", out.Pix)
	}
}

func workRGB(w, h int, pix []uint8) types.WorkRGB {
	return types.WorkRGB{W: w, H: h, Stride: w * 3, Pix: pix}
}

func BenchmarkTemporalBlend(b *testing.B) {
	bl := &TemporalBlend{}
	in := workRGB(160, 80, make([]uint8, 160*80*3))
	for i := range in.Pix {
		in.Pix[i] = byte(i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := bl.Blend(context.Background(), in, 0.3); err != nil {
			b.Fatal(err)
		}
	}
}
