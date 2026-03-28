package render

import (
	"context"
	"fmt"
	"video-terminal/types"
)

type BayerDither struct {
	bias4     [5][64]int
	bias8     [5][64]int
	biasReady bool
	luma      []uint8
}

func (d *BayerDither) Dither(ctx context.Context, in types.WorkRGB, preset types.Preset) (types.WorkRGB, error) {
	_ = ctx

	if in.W <= 0 || in.H <= 0 || len(in.Pix) < in.Stride*in.H {
		return types.WorkRGB{}, fmt.Errorf("invalid work buffer")
	}

	var matrixSize int
	var thresholdLookup func(x, y int) uint8
	var dynamic bool
	var minBias int
	var maxBias int
	var sourceLuma []uint8
	var biasTable *[5][64]int

	switch preset {
	case types.PresetQuality, types.PresetCRT:
		matrixSize = 8
		thresholdLookup = func(x, y int) uint8 {
			return bayer8x8[y%8][x%8]
		}
		dynamic = true
		minBias = 0
		maxBias = 4
		biasTable = d.ensureBiasTable(8)
	default:
		matrixSize = 4
		thresholdLookup = func(x, y int) uint8 {
			return bayer4x4[y%4][x%4]
		}
		dynamic = false
		minBias = 2
		maxBias = 2
		biasTable = d.ensureBiasTable(4)
	}

	if dynamic {
		sourceLuma = d.buildLumaBuffer(in)
	}

	out := in
	for y := 0; y < in.H; y++ {
		row := y * in.Stride
		tileY := y / matrixSize
		biasRange := maxBias
		lastTileX := -1
		for x := 0; x < in.W; x++ {
			idx := row + x*3
			threshold := thresholdLookup(x, y)
			if dynamic && maxBias > minBias {
				tileX := x / matrixSize
				if tileX != lastTileX {
					contrast := tileContrast(sourceLuma, in.W, in.H, tileX, tileY, matrixSize)
					biasRange = minBias + (contrast*(maxBias-minBias))/255
					lastTileX = tileX
				}
			}

			bias := (*biasTable)[biasRange][threshold]
			out.Pix[idx+0] = applyBias(out.Pix[idx+0], bias)
			out.Pix[idx+1] = applyBias(out.Pix[idx+1], bias)
			out.Pix[idx+2] = applyBias(out.Pix[idx+2], bias)
		}
	}

	return out, nil
}

func (d *BayerDither) buildLumaBuffer(in types.WorkRGB) []uint8 {
	need := in.W * in.H
	if cap(d.luma) < need {
		d.luma = make([]uint8, need)
	}
	buf := d.luma[:need]
	for y := 0; y < in.H; y++ {
		row := y * in.Stride
		for x := 0; x < in.W; x++ {
			idx := row + x*3
			buf[y*in.W+x] = lumaFromRGB(in.Pix[idx+0], in.Pix[idx+1], in.Pix[idx+2])
		}
	}
	return buf
}

func (d *BayerDither) ensureBiasTable(size int) *[5][64]int {
	if d == nil {
		if size == 8 {
			table := buildBiasTable(64, 4)
			return &table
		}
		table := buildBiasTable(16, 4)
		return &table
	}
	if !d.biasReady {
		d.bias4 = buildBiasTable(16, 4)
		d.bias8 = buildBiasTable(64, 4)
		d.biasReady = true
	}
	if size == 8 {
		return &d.bias8
	}
	return &d.bias4
}

func tileContrast(luma []uint8, w, h, tileX, tileY, tileSize int) int {
	if tileSize <= 0 {
		return 0
	}

	x0 := tileX * tileSize
	y0 := tileY * tileSize
	if x0 >= w || y0 >= h {
		return 0
	}

	x1 := minInt(x0+tileSize-1, w-1)
	y1 := minInt(y0+tileSize-1, h-1)
	mx := minInt(x0+tileSize/2, w-1)
	my := minInt(y0+tileSize/2, h-1)

	a := int(luma[y0*w+x0])
	b := int(luma[y0*w+x1])
	c := int(luma[y1*w+x0])
	d := int(luma[y1*w+x1])
	e := int(luma[my*w+mx])

	avg := (a + b + c + d + e) / 5
	return (absInt(a-avg) + absInt(b-avg) + absInt(c-avg) + absInt(d-avg) + absInt(e-avg)) / 5
}

func buildBiasTable(area, maxBias int) [5][64]int {
	var table [5][64]int
	if area <= 0 || maxBias <= 0 {
		return table
	}

	for biasRange := 0; biasRange <= maxBias && biasRange < len(table); biasRange++ {
		for threshold := 0; threshold < 64; threshold++ {
			table[biasRange][threshold] = threshold*((biasRange*2)+1)/area - biasRange
		}
	}

	return table
}

func lumaFromRGB(r, g, b uint8) uint8 {
	// integer approximation of sRGB luma: 0.299R + 0.587G + 0.114B
	return uint8((77*int(r) + 150*int(g) + 29*int(b)) >> 8)
}

func applyBias(c uint8, bias int) uint8 {
	if bias == 0 {
		return c
	}

	value := int(c) + bias
	if value < 0 {
		value = 0
	} else if value > 255 {
		value = 255
	}

	return uint8(value)
}

func absInt(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
