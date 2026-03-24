package pipeline

import (
	"context"
	"video-terminal/types"
)

type Decoder interface {
	Next(ctx context.Context) (types.FrameRGB, error)
}

type Resizer interface {
	Resize(ctx context.Context, src types.FrameRGB, termW, termH int) (types.WorkRGB, error)
}

type Quantizer interface {
	Quantize(ctx context.Context, in types.WorkRGB, mode types.ColorMode) (types.WorkRGB, error)
}

type Dither interface {
	Dither(ctx context.Context, in types.WorkRGB, preset types.Preset) (types.WorkRGB, error)
}

type Mapper interface {
	Map(ctx context.Context, in types.WorkRGB) (types.CellGrid, error)
}

type MapperInto interface {
	MapInto(ctx context.Context, in types.WorkRGB, dst *types.CellGrid) error
}

type Differ interface {
	Diff(ctx context.Context, curr types.CellGrid, prev *types.CellGrid) ([]types.DiffOp, error)
}

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
