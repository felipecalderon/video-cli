package diff

import (
	"context"
	"testing"
	"video-terminal/types"
)

func TestByteDifferGroupsContiguousRuns(t *testing.T) {
	grid := types.CellGrid{
		W: 4,
		H: 1,
		Cells: []types.Cell{
			{Top: [3]uint8{1, 2, 3}, Bottom: [3]uint8{4, 5, 6}, Ch: '▀'},
			{Top: [3]uint8{1, 2, 3}, Bottom: [3]uint8{4, 5, 6}, Ch: '▀'},
			{Top: [3]uint8{7, 8, 9}, Bottom: [3]uint8{10, 11, 12}, Ch: '▀'},
			{Top: [3]uint8{7, 8, 9}, Bottom: [3]uint8{10, 11, 12}, Ch: '▀'},
		},
	}

	ops, err := ByteDiffer{}.Diff(context.Background(), grid, nil)
	if err != nil {
		t.Fatalf("Diff returned error: %v", err)
	}

	if got, want := len(ops), 2; got != want {
		t.Fatalf("unexpected op count: got %d want %d", got, want)
	}
	if got, want := ops[0].Text, "▀▀"; got != want {
		t.Fatalf("unexpected first span text: got %q want %q", got, want)
	}
	if got, want := ops[1].Text, "▀▀"; got != want {
		t.Fatalf("unexpected second span text: got %q want %q", got, want)
	}
}
