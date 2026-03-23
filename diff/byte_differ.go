package diff

import (
	"context"
	"fmt"
	"video-terminal/types"
)

type ByteDiffer struct{}

func (ByteDiffer) Diff(ctx context.Context, curr types.CellGrid, prev *types.CellGrid) ([]types.DiffOp, error) {
	_ = ctx

	if curr.W <= 0 || curr.H <= 0 || len(curr.Cells) != curr.W*curr.H {
		return nil, fmt.Errorf("invalid current cell grid")
	}

	ops := make([]types.DiffOp, 0, curr.W*curr.H/4)

	for y := 0; y < curr.H; y++ {
		for x := 0; x < curr.W; x++ {
			idx := y*curr.W + x
			cc := curr.Cells[idx]

			changed := true
			if prev != nil && prev.W == curr.W && prev.H == curr.H && len(prev.Cells) == len(curr.Cells) {
				pc := prev.Cells[idx]
				changed = cc != pc
			}

			if changed {
				ops = append(ops, types.DiffOp{
					X:  x,
					Y:  y,
					FG: cc.Top,
					BG: cc.Bottom,
					Ch: cc.Ch,
				})
			}
		}
	}

	return ops, nil
}
