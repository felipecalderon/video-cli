package render

import (
	"context"
	"fmt"
	"video-terminal/types"
)

const upperHalfBlock = rune(0x2580)

type BlockMapper struct{}

func (BlockMapper) Map(ctx context.Context, in types.WorkRGB) (types.CellGrid, error) {
	var grid types.CellGrid
	if err := (BlockMapper{}).MapInto(ctx, in, &grid); err != nil {
		return types.CellGrid{}, err
	}
	return grid, nil
}

func (BlockMapper) MapInto(ctx context.Context, in types.WorkRGB, dst *types.CellGrid) error {
	_ = ctx

	if in.W <= 0 || in.H <= 0 || in.H%2 != 0 || len(in.Pix) < in.Stride*in.H {
		return fmt.Errorf("invalid work buffer")
	}
	if dst == nil {
		return fmt.Errorf("nil destination grid")
	}

	gridH := in.H / 2
	size := in.W * gridH
	if cap(dst.Cells) < size {
		dst.Cells = make([]types.Cell, size)
	} else {
		dst.Cells = dst.Cells[:size]
	}
	dst.W = in.W
	dst.H = gridH

	for y := 0; y < gridH; y++ {
		topY := y * 2
		botY := topY + 1

		for x := 0; x < in.W; x++ {
			topIdx := topY*in.Stride + x*3
			botIdx := botY*in.Stride + x*3

			dst.Cells[y*in.W+x] = types.Cell{
				Top: [3]uint8{
					in.Pix[topIdx+0],
					in.Pix[topIdx+1],
					in.Pix[topIdx+2],
				},
				Bottom: [3]uint8{
					in.Pix[botIdx+0],
					in.Pix[botIdx+1],
					in.Pix[botIdx+2],
				},
				Ch: upperHalfBlock,
			}
		}
	}

	return nil
}
