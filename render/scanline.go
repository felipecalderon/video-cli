package render

import (
	"context"
	"fmt"
	"math"
	"video-terminal/types"
)

type ScanlineEffect struct{}

func (ScanlineEffect) Apply(ctx context.Context, in types.WorkRGB, preset types.Preset) (types.WorkRGB, error) {
	_ = ctx

	if in.W <= 0 || in.H <= 0 || len(in.Pix) < in.Stride*in.H {
		return types.WorkRGB{}, fmt.Errorf("invalid work buffer")
	}

	strength := 0.0
	phosphorLift := 0.0
	maskStrength := 0.0
	switch preset {
	case types.PresetCRT:
		strength = 0.34
		phosphorLift = 0.03
		maskStrength = 0.06
	case types.PresetQuality:
		strength = 0.12
		phosphorLift = 0.02
	default:
		return in, nil
	}

	out := in

	for y := 0; y < in.H; y++ {
		rowFactor := 1.0
		if y%2 == 1 {
			rowFactor -= strength
		}
		if rowFactor < 0.28 {
			rowFactor = 0.28
		}

		row := y * in.Stride
		for x := 0; x < in.W; x++ {
			idx := row + x*3
			maskPhase := (x + y) % 3
			for c := 0; c < 3; c++ {
				channelFactor := 1.0
				if maskStrength > 0 {
					switch maskPhase {
					case 0:
						if c == 0 {
							channelFactor += maskStrength
						} else {
							channelFactor -= maskStrength / 2
						}
					case 1:
						if c == 1 {
							channelFactor += maskStrength
						} else {
							channelFactor -= maskStrength / 2
						}
					case 2:
						if c == 2 {
							channelFactor += maskStrength
						} else {
							channelFactor -= maskStrength / 2
						}
					}
				}
				v := float64(out.Pix[idx+c])*rowFactor*channelFactor + phosphorLift*255
				if v > 255 {
					v = 255
				}
				out.Pix[idx+c] = uint8(math.Round(v))
			}
		}
	}

	return out, nil
}
