package diff

import (
	"context"
	"fmt"
	"video-terminal/types"
)

const (
	DefaultDiffThreshold = 3
	keyframeInterval     = 120
	refreshCycleFrames   = 30
)

type ByteDiffer struct {
	ops        []types.DiffOp
	frameCount int
	Threshold  int
}

func (d *ByteDiffer) Diff(ctx context.Context, curr types.CellGrid, prev *types.CellGrid) ([]types.DiffOp, error) {
	_ = ctx

	if curr.W <= 0 || curr.H <= 0 || len(curr.Cells) != curr.W*curr.H {
		return nil, fmt.Errorf("invalid current cell grid")
	}

	if d == nil {
		return nil, fmt.Errorf("nil differ")
	}

	d.frameCount++

	threshold := d.Threshold
	if threshold <= 0 {
		threshold = DefaultDiffThreshold
	}

	estimate := curr.W * curr.H / 4
	if cap(d.ops) < estimate {
		d.ops = make([]types.DiffOp, 0, estimate)
	}
	ops := d.ops[:0]

	isFullKeyframe := (d.frameCount % keyframeInterval == 0)
	rowsPerFrame := (curr.H + refreshCycleFrames - 1) / refreshCycleFrames
	if rowsPerFrame < 1 {
		rowsPerFrame = 1
	}
	startRefreshRow := (d.frameCount * rowsPerFrame) % curr.H
	endRefreshRow := startRefreshRow + rowsPerFrame

	for y := 0; y < curr.H; y++ {
		rowStart := -1
		var run types.Cell

		forceRowRefresh := isFullKeyframe || (y >= startRefreshRow && y < endRefreshRow) || (endRefreshRow > curr.H && y < (endRefreshRow%curr.H))

		flush := func(endX int) {
			if rowStart < 0 || endX <= rowStart {
				return
			}

			ops = append(ops, types.DiffOp{
				X:      rowStart,
				Y:      y,
				FG:     run.Top,
				BG:     run.Bottom,
				Ch:     run.Ch,
				RunLen: endX - rowStart,
			})
			rowStart = -1
		}

		for x := 0; x < curr.W; x++ {
			idx := y*curr.W + x
			cc := curr.Cells[idx]

			changed := true
			hasPrev := prev != nil && prev.W == curr.W && prev.H == curr.H && len(prev.Cells) == len(curr.Cells)
			if hasPrev {
				pc := prev.Cells[idx]
				if forceRowRefresh {
					changed = (cc != pc)
				} else {
					changed = !isCellSimilar(cc, pc, threshold)
				}
			}

			if !changed {
				flush(x)
				continue
			}

			if hasPrev {
				prev.Cells[idx] = cc
			}

			if rowStart == -1 {
				rowStart = x
				run = cc
				continue
			}

			if cc.Top == run.Top && cc.Bottom == run.Bottom && cc.Ch == run.Ch {
				continue
			}

			flush(x)
			rowStart = x
			run = cc
		}

		flush(curr.W)
	}

	d.ops = ops
	return ops, nil
}

func isCellSimilar(a, b types.Cell, threshold int) bool {
	if a.Ch != b.Ch {
		return false
	}
	return isColorSimilar(a.Top, b.Top, threshold) && isColorSimilar(a.Bottom, b.Bottom, threshold)
}

func isColorSimilar(a, b [3]uint8, threshold int) bool {
	if a == b {
		return true
	}
	if threshold <= 0 {
		return false
	}

	dr := absDiff(a[0], b[0])
	dg := absDiff(a[1], b[1])
	db := absDiff(a[2], b[2])

	lumaDiff := (77*dr + 150*dg + 29*db) >> 8
	maxDiff := dr
	if dg > maxDiff {
		maxDiff = dg
	}
	if db > maxDiff {
		maxDiff = db
	}

	return lumaDiff <= threshold && maxDiff <= threshold*2
}

func absDiff(x, y uint8) int {
	if x > y {
		return int(x - y)
	}
	return int(y - x)
}
