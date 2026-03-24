package pipeline

import (
	"context"
	"errors"
	"io"
	"time"
	"video-terminal/types"
)

var errNilStage = errors.New("pipeline stage is nil")

func (p Pipeline) Run(ctx context.Context, params types.PipelineParams) error {
	if p.Decoder == nil || p.Resizer == nil || p.Quantizer == nil || p.Dither == nil || p.Mapper == nil || p.Differ == nil || p.Output == nil {
		return errNilStage
	}

	frameDuration := time.Second / 15
	if params.FpsTarget > 0 {
		frameDuration = time.Second / time.Duration(params.FpsTarget)
	}

	var prev *types.CellGrid
	var buffers [2]types.CellGrid
	var currIdx int
	mapperInto, supportsReuse := p.Mapper.(MapperInto)
	nextTick := time.Now().Add(frameDuration)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		start := time.Now()

		frame, err := p.Decoder.Next(ctx)
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return err
		}

		work, err := p.Resizer.Resize(ctx, frame, params.TermW, params.TermH)
		if err != nil {
			return err
		}

		if p.Temporal != nil && params.BlendAlpha > 0 {
			blended, err := p.Temporal.Blend(ctx, work, params.BlendAlpha)
			if err != nil {
				return err
			}
			work = blended
		}

		dithered, err := p.Dither.Dither(ctx, work, params.Preset)
		if err != nil {
			return err
		}

		quantized, err := p.Quantizer.Quantize(ctx, dithered, params.ColorMode)
		if err != nil {
			return err
		}

		var grid types.CellGrid
		if supportsReuse {
			curr := &buffers[currIdx]
			if err := mapperInto.MapInto(ctx, quantized, curr); err != nil {
				return err
			}
			grid = *curr
		} else {
			mapped, err := p.Mapper.Map(ctx, quantized)
			if err != nil {
				return err
			}
			grid = mapped
		}

		ops, err := p.Differ.Diff(ctx, grid, prev)
		if err != nil {
			return err
		}

		if err := p.Output.Write(ctx, ops); err != nil {
			return err
		}

		if supportsReuse {
			prev = &buffers[currIdx]
			currIdx = 1 - currIdx
		} else {
			prevFrame := grid
			prev = &prevFrame
		}

		elapsed := time.Since(start)
		targetTick := nextTick
		nextTick = nextTick.Add(frameDuration)

		if elapsed < frameDuration {
			sleep := time.Until(targetTick)
			if sleep > 0 {
				timer := time.NewTimer(sleep)
				select {
				case <-ctx.Done():
					timer.Stop()
					return ctx.Err()
				case <-timer.C:
				}
			}
		} else if time.Since(targetTick) > frameDuration {

			nextTick = time.Now().Add(frameDuration)
		}
	}
}
