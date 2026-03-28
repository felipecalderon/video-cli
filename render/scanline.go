package render

import (
	"context"
	"fmt"
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
	const scale = 1024
	const denom = scale * scale
	minRow := (28*scale + 50) / 100
	baseRow := scale
	oddRow := int((1.0-strength)*scale + 0.5)
	if oddRow < minRow {
		oddRow = minRow
	}
	liftAdd := int(phosphorLift*255 + 0.5)
	maskUp := scale
	maskDown := scale
	if maskStrength > 0 {
		maskUp = int((1.0+maskStrength)*scale + 0.5)
		maskDown = int((1.0-maskStrength/2.0)*scale + 0.5)
	}

	for y := 0; y < in.H; y++ {
		rowFactor := baseRow
		if y%2 == 1 {
			rowFactor = oddRow
		}

		row := y * in.Stride
		for x := 0; x < in.W; x++ {
			idx := row + x*3
			maskPhase := (x + y) % 3
			for c := 0; c < 3; c++ {
				channelFactor := scale
				if maskStrength > 0 {
					switch maskPhase {
					case 0:
						if c == 0 {
							channelFactor = maskUp
						} else {
							channelFactor = maskDown
						}
					case 1:
						if c == 1 {
							channelFactor = maskUp
						} else {
							channelFactor = maskDown
						}
					case 2:
						if c == 2 {
							channelFactor = maskUp
						} else {
							channelFactor = maskDown
						}
					}
				}
				v := int(out.Pix[idx+c])
				mix := v * rowFactor * channelFactor
				mix = (mix + denom/2) / denom
				mix += liftAdd
				if mix > 255 {
					mix = 255
				} else if mix < 0 {
					mix = 0
				}
				out.Pix[idx+c] = uint8(mix)
			}
		}
	}

	return out, nil
}
