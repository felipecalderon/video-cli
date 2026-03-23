package render

import (
	"context"
	"fmt"
	"video-terminal/types"
)

type NearestResizer struct{}

func (NearestResizer) Resize(ctx context.Context, src types.FrameRGB, termW, termH int) (types.WorkRGB, error) {
	_ = ctx

	if src.W <= 0 || src.H <= 0 || len(src.Pix) < src.Stride*src.H {
		return types.WorkRGB{}, fmt.Errorf("invalid source frame")
	}
	if termW <= 0 || termH <= 0 {
		return types.WorkRGB{}, fmt.Errorf("invalid terminal size")
	}

	workH := termH * 2
	stride := termW * 3
	pix := make([]byte, stride*workH)

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
