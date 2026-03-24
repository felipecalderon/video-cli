package render

import (
	"context"
	"fmt"
	"math"
	"video-terminal/types"
)

type TemporalBlend struct {
	prev    types.WorkRGB
	hasPrev bool
}

func (t *TemporalBlend) Blend(ctx context.Context, in types.WorkRGB, alpha float64) (types.WorkRGB, error) {
	_ = ctx

	if in.W <= 0 || in.H <= 0 || len(in.Pix) < in.Stride*in.H {
		return types.WorkRGB{}, fmt.Errorf("invalid work buffer")
	}

	if alpha <= 0 {
		t.storePrev(in)
		return in, nil
	}

	if alpha > 1 {
		alpha = 1
	}

	if !t.hasPrev || !sameWorkSize(t.prev, in) || len(t.prev.Pix) < in.Stride*in.H {
		t.initPrev(in)
		return in, nil
	}

	mix := uint16(math.Round(alpha * 255))
	if mix == 0 {
		t.storePrev(in)
		return in, nil
	}
	inv := uint16(255) - mix

	pix := in.Pix
	prev := t.prev.Pix
	max := in.Stride * in.H
	for i := 0; i < max; i += 3 {
		pix[i] = blendChannel(pix[i], prev[i], inv, mix)
		pix[i+1] = blendChannel(pix[i+1], prev[i+1], inv, mix)
		pix[i+2] = blendChannel(pix[i+2], prev[i+2], inv, mix)
	}

	copy(prev, pix)
	return in, nil
}

func (t *TemporalBlend) initPrev(in types.WorkRGB) {
	t.prev = types.WorkRGB{W: in.W, H: in.H, Stride: in.Stride, Pix: make([]uint8, in.Stride*in.H)}
	copy(t.prev.Pix, in.Pix)
	t.hasPrev = true
}

func (t *TemporalBlend) storePrev(in types.WorkRGB) {
	if !t.hasPrev || !sameWorkSize(t.prev, in) || len(t.prev.Pix) < in.Stride*in.H {
		t.initPrev(in)
		return
	}
	copy(t.prev.Pix, in.Pix)
}

func sameWorkSize(a, b types.WorkRGB) bool {
	return a.W == b.W && a.H == b.H && a.Stride == b.Stride
}

func blendChannel(curr, prev uint8, inv, mix uint16) uint8 {
	v := uint16(curr)*inv + uint16(prev)*mix + 127
	return uint8(v / 255)
}
