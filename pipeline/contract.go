package pipeline

import (
	"context"
	"video-terminal/types"
)

// Decoder produces source frames (RGB24) from an input stream.
type Decoder interface {
	Next(ctx context.Context) (types.FrameRGB, error)
}

// Resizer maps source frames to terminal resolution (with subcells).
type Resizer interface {
	Resize(ctx context.Context, src types.FrameRGB, termW, termH int) (types.WorkRGB, error)
}

// Quantizer reduces color precision based on color mode.
type Quantizer interface {
	Quantize(ctx context.Context, in types.WorkRGB, mode types.ColorMode) (types.WorkRGB, error)
}

// Dither applies spatial dithering.
type Dither interface {
	Dither(ctx context.Context, in types.WorkRGB, preset types.Preset) (types.WorkRGB, error)
}

// Mapper converts work pixels to terminal cells.
type Mapper interface {
	Map(ctx context.Context, in types.WorkRGB) (types.CellGrid, error)
}

// Differ computes the minimal set of cell updates.
type Differ interface {
	Diff(ctx context.Context, curr types.CellGrid, prev *types.CellGrid) ([]types.DiffOp, error)
}

// Output emits ANSI commands to stdout (or any writer).
type Output interface {
	Write(ctx context.Context, ops []types.DiffOp) error
}

type Pipeline struct {
	Decoder   Decoder
	Resizer   Resizer
	Quantizer Quantizer
	Dither    Dither
	Mapper    Mapper
	Differ    Differ
	Output    Output
}
