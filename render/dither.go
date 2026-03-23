package render

import (
	"context"
	"fmt"
	"video-terminal/types"
)

type BayerDither struct{}

func (BayerDither) Dither(ctx context.Context, in types.WorkRGB, preset types.Preset) (types.WorkRGB, error) {
	_ = ctx

	if in.W <= 0 || in.H <= 0 || len(in.Pix) < in.Stride*in.H {
		return types.WorkRGB{}, fmt.Errorf("invalid work buffer")
	}

	var matrixSize int
	var thresholdLookup func(x, y int) uint8

	switch preset {
	case types.PresetQuality, types.PresetCRT:
		matrixSize = 8
		thresholdLookup = func(x, y int) uint8 {
			return bayer8x8[y%8][x%8]
		}
	default:
		matrixSize = 4
		thresholdLookup = func(x, y int) uint8 {
			return bayer4x4[y%4][x%4]
		}
	}

	area := matrixSize * matrixSize
	out := in
	for y := 0; y < in.H; y++ {
		row := y * in.Stride
		for x := 0; x < in.W; x++ {
			idx := row + x*3
			threshold := thresholdLookup(x, y)
			out.Pix[idx+0] = ditherChannel(in.Pix[idx+0], threshold, area)
			out.Pix[idx+1] = ditherChannel(in.Pix[idx+1], threshold, area)
			out.Pix[idx+2] = ditherChannel(in.Pix[idx+2], threshold, area)
		}
	}

	return out, nil
}
