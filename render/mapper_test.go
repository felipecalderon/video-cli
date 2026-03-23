package render

import (
	"context"
	"testing"
	"video-terminal/types"
)

func TestBlockMapperMapIntoReusesBuffer(t *testing.T) {
	in := types.WorkRGB{
		W:      2,
		H:      4,
		Stride: 6,
		Pix: []byte{
			1, 2, 3, 4, 5, 6,
			7, 8, 9, 10, 11, 12,
			13, 14, 15, 16, 17, 18,
			19, 20, 21, 22, 23, 24,
		},
	}

	var grid types.CellGrid
	mapper := BlockMapper{}
	if err := mapper.MapInto(context.Background(), in, &grid); err != nil {
		t.Fatalf("MapInto returned error: %v", err)
	}

	firstCap := cap(grid.Cells)
	if firstCap == 0 {
		t.Fatalf("expected allocated cells buffer")
	}

	if err := mapper.MapInto(context.Background(), in, &grid); err != nil {
		t.Fatalf("MapInto returned error on reuse: %v", err)
	}

	if got := cap(grid.Cells); got != firstCap {
		t.Fatalf("expected capacity to be reused, got %d want %d", got, firstCap)
	}
}
