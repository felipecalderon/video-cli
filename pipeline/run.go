package pipeline

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
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

	// Phase 3: Initialize SlotPool
	// Using size 3: 1 being filled by prefetcher, 1 being processed, 1 being written/released
	pool := NewSlotPool(3)
	
	// Phase 4A: Prefetcher Channel
	// Depth 2 to allow smooth prefetching without getting too far ahead of audio
	prefetchChan := make(chan *Slot, 2)
	errChan := make(chan error, 1)

	// Goroutine prefetcher (Phase 4A)
	go func() {
		defer close(prefetchChan)
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}

			slot := pool.Acquire()
			// We use NextInto to avoid internal allocations in the decoder
			if err := p.Decoder.NextInto(ctx, &slot.Frame); err != nil {
				pool.Release(slot)
				if !errors.Is(err, io.EOF) && !errors.Is(err, context.Canceled) {
					errChan <- err
				}
				return
			}
			
			select {
			case <-ctx.Done():
				pool.Release(slot)
				return
			case prefetchChan <- slot:
			}
		}
	}()

	var prev *types.CellGrid
	mapperInto, supportsReuse := p.Mapper.(MapperInto)

	framesParsed := 0
	framesSkippedTotal := 0
	var syncTimer *time.Timer
	defer func() {
		if syncTimer != nil {
			syncTimer.Stop()
		}
	}()

	var timers PipelineTimers
	var lastProfileUpdate time.Time
	var profileHUD string

	for {
		loopStart := time.Now()

		select {
		case <-ctx.Done():
			return ctx.Err()
		case err := <-errChan:
			return err
		default:
		}

		select {
		case newSize := <-params.ResizeChan:
			params.TermW = newSize[0]
			params.TermH = newSize[1]
			_ = p.Output.Clear(ctx)
			prev = nil
		default:
		}

		// --- DECISION DE SINCRONIA ---
		pts := time.Duration(framesParsed) * frameDuration
		
		tSyncStart := time.Now()
		if params.Clock != nil {
			at := params.Clock.CurrentTime()
			diff := pts - at
			
			if diff > 10*time.Millisecond {
				if syncTimer == nil {
					syncTimer = time.NewTimer(diff)
				} else {
					if !syncTimer.Stop() {
						select {
						case <-syncTimer.C:
						default:
						}
					}
					syncTimer.Reset(diff)
				}
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-syncTimer.C:
				}
			}
			
			if diff < -25*time.Millisecond {
				skipCount := 0
				for diff < -25*time.Millisecond && skipCount < params.FpsTarget {
					// NOTE: When skipping, we consume from prefetchChan
					select {
					case slot, ok := <-prefetchChan:
						if !ok {
							break
						}
						pool.Release(slot)
					default:
						// If prefetch is empty, we might need to wait or just break
						// For now, let's just break to not block the sync loop indefinitely
						goto skipEnd 
					}
					framesParsed++
					framesSkippedTotal++
					skipCount++
					
					at = params.Clock.CurrentTime()
					pts = time.Duration(framesParsed) * frameDuration
					diff = pts - at
				}
			skipEnd:
				if diff < -25*time.Millisecond {
					timers.SyncWait.Mark(time.Since(tSyncStart))
					continue
				}
			}
		} else {
			time.Sleep(frameDuration)
		}
		timers.SyncWait.Mark(time.Since(tSyncStart))

		// Consume from prefetcher
		var slot *Slot
		select {
		case <-ctx.Done():
			return ctx.Err()
		case err := <-errChan:
			return err
		case s, ok := <-prefetchChan:
			if !ok {
				return nil // EOF
			}
			slot = s
		}

		tDecode := time.Now()
		// (Measurement is now just the overhead of the channel + sync, but we mark it)
		timers.Decode.Mark(time.Since(tDecode))
		
		framesParsed++

		// Processing using Slot buffers
		tResize := time.Now()
		work, err := p.Resizer.Resize(ctx, slot.Frame, params.TermW, params.TermH)
		timers.Resize.Mark(time.Since(tResize))
		if err != nil {
			pool.Release(slot)
			return err
		}

		if p.Temporal != nil && params.BlendAlpha > 0 {
			tTemp := time.Now()
			// We could optimize Blend to use slot.WorkTemp, but for now we keep it simple
			blended, err := p.Temporal.Blend(ctx, work, params.BlendAlpha)
			timers.TemporalBlend.Mark(time.Since(tTemp))
			if err != nil {
				pool.Release(slot)
				return err
			}
			work = blended
		}

		if p.Scanliner != nil {
			tScan := time.Now()
			scanned, err := p.Scanliner.Apply(ctx, work, params.Preset)
			timers.Scanline.Mark(time.Since(tScan))
			if err != nil {
				pool.Release(slot)
				return err
			}
			work = scanned
		}

		tQuant := time.Now()
		quantized, err := p.Quantizer.Quantize(ctx, work, params.ColorMode)
		timers.Quantize.Mark(time.Since(tQuant))
		if err != nil {
			pool.Release(slot)
			return err
		}

		tDither := time.Now()
		dithered, err := p.Dither.Dither(ctx, quantized, params.Preset)
		timers.Dither.Mark(time.Since(tDither))
		if err != nil {
			pool.Release(slot)
			return err
		}

		tMap := time.Now()
		if supportsReuse {
			if err := mapperInto.MapInto(ctx, dithered, &slot.Grid); err != nil {
				pool.Release(slot)
				return err
			}
		} else {
			mapped, err := p.Mapper.Map(ctx, dithered)
			if err != nil {
				pool.Release(slot)
				return err
			}
			slot.Grid = mapped
		}
		timers.Map.Mark(time.Since(tMap))

		tDiff := time.Now()
		ops, err := p.Differ.Diff(ctx, slot.Grid, prev)
		timers.Diff.Mark(time.Since(tDiff))
		if err != nil {
			pool.Release(slot)
			return err
		}

		if params.Profile && params.TermH > 0 {
			if time.Since(lastProfileUpdate) > time.Second/4 {
				decP50 := float64(timers.Decode.Metrics().P50.Microseconds()) / 1000.0
				dithP50 := float64(timers.Dither.Metrics().P50.Microseconds()) / 1000.0
				outP50 := float64(timers.Output.Metrics().P50.Microseconds()) / 1000.0
				syncP50 := float64(timers.SyncWait.Metrics().P50.Microseconds()) / 1000.0
				mapP50 := float64(timers.Map.Metrics().P50.Microseconds()) / 1000.0

				profileHUD = fmt.Sprintf("fps:%d dec:%.1fms dith:%.1fms map:%.1fms out:%.1fms sync:%.1fms skip:%d",
					params.FpsTarget, decP50, dithP50, mapP50, outP50, syncP50, framesSkippedTotal)
				lastProfileUpdate = time.Now()

				if params.ProfileLog != "" {
					go func(logMsg string) {
						f, err := os.OpenFile(params.ProfileLog, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
						if err == nil {
							f.WriteString(time.Now().Format(time.RFC3339) + " " + logMsg + "\n")
							f.Close()
						}
					}(profileHUD)
				}
			}
			if profileHUD != "" {
				ops = append(ops, types.DiffOp{
					X: 0, Y: params.TermH - 1,
					FG: [3]uint8{255, 255, 0}, BG: [3]uint8{0, 0, 0},
					Text: []rune(profileHUD),
				})
			}
		}

		tOut := time.Now()
		if err := p.Output.Write(ctx, ops); err != nil {
			pool.Release(slot)
			return err
		}
		timers.Output.Mark(time.Since(tOut))

		// Update prev before releasing slot
		// Need a copy of the grid if we don't want to hold the slot forever
		// But in sequential logic, we can just point to it.
		// Wait, if we release the slot, the grid might be reused!
		// So we must COPY the grid or HOLD two slots.
		// For simplicity, let's copy the grid to a persistent buffer in prev.
		if prev == nil {
			prev = &types.CellGrid{}
		}
		copyGrid(prev, &slot.Grid)

		pool.Release(slot)
		timers.FrameTotal.Mark(time.Since(loopStart))
	}
}

func copyGrid(dst, src *types.CellGrid) {
	dst.W = src.W
	dst.H = src.H
	if cap(dst.Cells) < len(src.Cells) {
		dst.Cells = make([]types.Cell, len(src.Cells))
	}
	dst.Cells = dst.Cells[:len(src.Cells)]
	copy(dst.Cells, src.Cells)
}
