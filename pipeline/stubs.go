package pipeline

import (
	"context"
	"errors"
	"video-terminal/types"
)

var ErrNotImplemented = errors.New("not implemented")

type NullDecoder struct{}

func (NullDecoder) Next(ctx context.Context) (types.FrameRGB, error) {
	_ = ctx
	return types.FrameRGB{}, ErrNotImplemented
}

type NearestResizer struct{}

func (NearestResizer) Resize(ctx context.Context, src types.FrameRGB, termW, termH int) (types.WorkRGB, error) {
	_ = ctx
	_ = src
	_ = termW
	_ = termH
	return types.WorkRGB{}, ErrNotImplemented
}

type NoopQuantizer struct{}

func (NoopQuantizer) Quantize(ctx context.Context, in types.WorkRGB, mode types.ColorMode) (types.WorkRGB, error) {
	_ = ctx
	_ = mode
	return in, ErrNotImplemented
}

type NoopDither struct{}

func (NoopDither) Dither(ctx context.Context, in types.WorkRGB, preset types.Preset) (types.WorkRGB, error) {
	_ = ctx
	_ = preset
	return in, ErrNotImplemented
}

type NoopTemporal struct{}

func (NoopTemporal) Blend(ctx context.Context, in types.WorkRGB, alpha float64) (types.WorkRGB, error) {
	_ = ctx
	_ = alpha
	return in, ErrNotImplemented
}

type BlockMapper struct{}

func (BlockMapper) Map(ctx context.Context, in types.WorkRGB) (types.CellGrid, error) {
	_ = ctx
	_ = in
	return types.CellGrid{}, ErrNotImplemented
}

type ByteDiffer struct{}

func (ByteDiffer) Diff(ctx context.Context, curr types.CellGrid, prev *types.CellGrid) ([]types.DiffOp, error) {
	_ = ctx
	_ = curr
	_ = prev
	return nil, ErrNotImplemented
}

type StdoutOutput struct{}

func (StdoutOutput) Write(ctx context.Context, ops []types.DiffOp) error {
	_ = ctx
	_ = ops
	return ErrNotImplemented
}

func (StdoutOutput) Clear(ctx context.Context) error {
	_ = ctx
	return ErrNotImplemented
}
