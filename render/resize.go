package render

import (
	"context"
	"fmt"
	"video-terminal/types"
)

type NearestResizer struct {
	buf    []byte
	width  int
	height int
	stride int
}

func (r *NearestResizer) Resize(ctx context.Context, src types.FrameRGB, termW, termH int) (types.WorkRGB, error) {
	_ = ctx

	if src.W <= 0 || src.H <= 0 || len(src.Pix) < src.Stride*src.H {
		return types.WorkRGB{}, fmt.Errorf("invalid source frame")
	}
	if termW <= 0 || termH <= 0 {
		return types.WorkRGB{}, fmt.Errorf("invalid terminal size")
	}

	workH := termH * 2
	stride := termW * 3
	need := stride * workH
	if r == nil {
		return types.WorkRGB{}, fmt.Errorf("nil resizer")
	}
	if r.width != termW || r.height != workH || r.stride != stride || cap(r.buf) < need {
		r.buf = make([]byte, need)
		r.width = termW
		r.height = workH
		r.stride = stride
	}
	pix := r.buf[:need]

	for y := 0; y < workH; y++ {
		srcY := y * src.H / workH
		for x := 0; x < termW; x++ {
			srcX := x * src.W / termW

			srcIdx := srcY*src.Stride + srcX*3
			dstIdx := y*stride + x*3

			pix[dstIdx+0] = src.Pix[srcIdx+0]
			pix[dstIdx+1] = src.Pix[srcIdx+1]
			pix[dstIdx+2] = src.Pix[srcIdx+2]
		}
	}

	return types.WorkRGB{W: termW, H: workH, Stride: stride, Pix: pix}, nil
}
