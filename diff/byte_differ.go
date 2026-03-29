package diff

import (
	"context"
	"fmt"
	"video-terminal/types"
)

type ByteDiffer struct {
	ops   []types.DiffOp
	runes []rune
}

func (d *ByteDiffer) Diff(ctx context.Context, curr types.CellGrid, prev *types.CellGrid) ([]types.DiffOp, error) {
	_ = ctx

	if curr.W <= 0 || curr.H <= 0 || len(curr.Cells) != curr.W*curr.H {
		return nil, fmt.Errorf("invalid current cell grid")
	}

	if d == nil {
		return nil, fmt.Errorf("nil differ")
	}
	estimate := curr.W * curr.H / 4
	if cap(d.ops) < estimate {
		d.ops = make([]types.DiffOp, 0, estimate)
	}
	ops := d.ops[:0]
	estimateRunes := curr.W * curr.H
	if cap(d.runes) < estimateRunes {
		d.runes = make([]rune, 0, estimateRunes)
	}
	d.runes = d.runes[:0]

	for y := 0; y < curr.H; y++ {
		rowStart := -1
		var run types.Cell
		runStartIdx := len(d.runes)

		flush := func(endX int) {
			if rowStart < 0 || len(d.runes)-runStartIdx <= 0 {
				return
			}

			ops = append(ops, types.DiffOp{
				X:    rowStart,
				Y:    y,
				FG:   run.Top,
				BG:   run.Bottom,
				Ch:   run.Ch,
				Text: d.runes[runStartIdx:],
			})
			rowStart = -1
			runStartIdx = len(d.runes)
			_ = endX
		}

		for x := 0; x < curr.W; x++ {
			idx := y*curr.W + x
			cc := curr.Cells[idx]

			changed := true
			if prev != nil && prev.W == curr.W && prev.H == curr.H && len(prev.Cells) == len(curr.Cells) {
				pc := prev.Cells[idx]
				changed = cc != pc
			}

			if !changed {
				flush(x)
				continue
			}

			if rowStart == -1 {
				rowStart = x
				run = cc
				d.runes = append(d.runes, cc.Ch)
				continue
			}

			if cc.Top == run.Top && cc.Bottom == run.Bottom && x == rowStart+(len(d.runes)-runStartIdx) {
				d.runes = append(d.runes, cc.Ch)
				continue
			}

			flush(x)
			rowStart = x
			run = cc
			d.runes = append(d.runes, cc.Ch)
		}

		flush(curr.W)
	}

	d.ops = ops
	return ops, nil
}
