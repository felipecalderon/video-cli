package render

import (
	"context"
	"fmt"
	"video-terminal/types"
)

type ChannelQuantizer struct{}

func (ChannelQuantizer) Quantize(ctx context.Context, in types.WorkRGB, mode types.ColorMode) (types.WorkRGB, error) {
	_ = ctx

	if in.W <= 0 || in.H <= 0 || len(in.Pix) < in.Stride*in.H {
		return types.WorkRGB{}, fmt.Errorf("invalid work buffer")
	}

	levels := quantizeTruecolorLevels
	if mode == types.Color256 {
		levels = quantizeAnsi256Levels
	}

	out := in
	for y := 0; y < in.H; y++ {
		row := y * in.Stride
		for x := 0; x < in.W; x++ {
			idx := row + x*3
			out.Pix[idx+0] = quantizeChannel(in.Pix[idx+0], levels)
			out.Pix[idx+1] = quantizeChannel(in.Pix[idx+1], levels)
			out.Pix[idx+2] = quantizeChannel(in.Pix[idx+2], levels)
		}
	}

	return out, nil
}
