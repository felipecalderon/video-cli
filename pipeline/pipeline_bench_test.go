package pipeline

import (
	"context"
	"io"
	"testing"
	"video-terminal/types"
)

type mockDecoder struct {
	frame types.FrameRGB
}

func (m *mockDecoder) Next(ctx context.Context) (types.FrameRGB, error) {
	return m.frame, nil
}

func (m *mockDecoder) NextInto(ctx context.Context, frame *types.FrameRGB) error {
	if cap(frame.Pix) < len(m.frame.Pix) {
		frame.Pix = make([]byte, len(m.frame.Pix))
	}
	frame.Pix = frame.Pix[:len(m.frame.Pix)]
	copy(frame.Pix, m.frame.Pix)
	frame.W = m.frame.W
	frame.H = m.frame.H
	frame.Stride = m.frame.Stride
	return nil
}

type mockOutput struct{}

func (m *mockOutput) Write(ctx context.Context, ops []types.DiffOp) error { return nil }
func (m *mockOutput) Clear(ctx context.Context) error                   { return nil }

type mockResizer struct{}

func (m *mockResizer) Resize(ctx context.Context, src types.FrameRGB, termW, termH int) (types.WorkRGB, error) {
	return types.WorkRGB{W: termW, H: termH, Stride: termW * 3, Pix: make([]byte, termW*termH*3)}, nil
}

type mockQuantizer struct{}

func (m *mockQuantizer) Quantize(ctx context.Context, in types.WorkRGB, mode types.ColorMode) (types.WorkRGB, error) {
	return in, nil
}

type mockDither struct{}

func (m *mockDither) Dither(ctx context.Context, in types.WorkRGB, preset types.Preset) (types.WorkRGB, error) {
	return in, nil
}

type mockMapper struct{}

func (m *mockMapper) Map(ctx context.Context, in types.WorkRGB) (types.CellGrid, error) {
	return types.CellGrid{W: in.W / 2, H: in.H / 4, Cells: make([]types.Cell, (in.W/2)*(in.H/4))}, nil
}

type mockDiffer struct{}

func (m *mockDiffer) Diff(ctx context.Context, curr types.CellGrid, prev *types.CellGrid) ([]types.DiffOp, error) {
	return nil, nil
}

func BenchmarkPipelineProcessing(b *testing.B) {
	w, h := 160, 45
	frame := types.FrameRGB{
		W:      1920,
		H:      1080,
		Stride: 1920 * 3,
		Pix:    make([]byte, 1920*1080*3),
	}

	p := Pipeline{
		Decoder:   &mockDecoder{frame: frame},
		Resizer:   &mockResizer{},
		Quantizer: &mockQuantizer{},
		Dither:    &mockDither{},
		Mapper:    &mockMapper{},
		Differ:    &mockDiffer{},
		Output:    &mockOutput{},
	}

	params := types.PipelineParams{
		TermW:     w,
		TermH:     h,
		FpsTarget: 60, // High target to avoid sleeping in sync
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// We can't easily benchmark Run because it's an infinite loop.
	// But we can benchmark the internal logic or run it for a limited time.
	// For now, let's just test that it builds and runs.
	
	b.ResetTimer()
	b.ReportAllocs()
	
	// Since Run is infinite, we'd need a way to stop it after N frames.
	// Let's modify the benchmark to run a version of the loop.
	// (Simulated loop based on Run logic)
}
