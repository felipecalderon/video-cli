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

	// Iniciar contadores para sincronía
	framesParsed := 0
	var syncTimer *time.Timer
	defer func() {
		if syncTimer != nil {
			syncTimer.Stop()
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
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
		// El "Presentation Timestamp" (PTS) calculado según el número del frame.
		pts := time.Duration(framesParsed) * frameDuration
		
		// Si tenemos un reloj externo (Audio), sincronizamos.
		if params.Clock != nil {
			at := params.Clock.CurrentTime()
			diff := pts - at
			
			// Caso 1: Vamos muy rápido (el video está adelantado > 10ms)
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
					// Alcanzamos el punto de renderizado
				}
			}
			
			// Caso 2: Vamos muy lento (el video está atrasado > 25ms)
			// Saltamos frames de forma agresiva hasta alcanzar al audio.
			if diff < -25*time.Millisecond {
				skipCount := 0
				// Consumir frames hasta sincronizar o hasta un máximo de 1 segundo de "skip" por ciclo.
				for diff < -25*time.Millisecond && skipCount < params.FpsTarget {
					if _, err := p.Decoder.Next(ctx); err != nil {
						break 
					}
					framesParsed++
					skipCount++
					
					at = params.Clock.CurrentTime()
					pts = time.Duration(framesParsed) * frameDuration
					diff = pts - at
				}
				// Si aún después del skip sigue atrasado, continuamos el bucle principal inmediatamente.
				if diff < -25*time.Millisecond {
					continue
				}
			}
		} else {
			// Fallback: Si no hay audio, usamos un sleep controlado.
			time.Sleep(frameDuration)
		}

		frame, err := p.Decoder.Next(ctx)
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return err
		}
		framesParsed++

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

		if p.Scanliner != nil {
			scanned, err := p.Scanliner.Apply(ctx, work, params.Preset)
			if err != nil {
				return err
			}
			work = scanned
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
	}
}

