# Architecture — Initial Draft

## Goal
Define modules, boundaries, and data flow to avoid implicit decisions once implementation starts.

---

## Proposed modules

- `ingest/` (FFmpeg + video input)
- `pipeline/` (stage-based processing)
- `render/` (cell mapping + ANSI)
- `diff/` (difference engine)
- `term/` (capability detection + sizing)
- `timing/` (FPS control + synchronization)
- `profile/` (stats, metrics, profiling)

---

## Logical pipeline (stages)

1. Decode (RGB frames)
2. Resize to terminal resolution
3. Color conversion / linearization (if applicable)
4. Channel split
5. Per-channel quantization
6. Per-channel dithering
7. Map to cells (Unicode + ANSI)
8. Diff (current frame vs previous)
9. Output (stdout)

---

## Concurrency (pending decision)

- Pipeline-per-stage with Go channels?
- Lock-free double buffering?
- Backpressure handling if stdout is slow?

---

## FFmpeg integration (pending decision)

- Input: `ffmpeg` -> `rawvideo` over stdout
- Formats: `rgb24` / `rgba` / `bgr24`
- Synchronization: external timestamps or fixed by target FPS?

---

## MVP success criteria (measurable)

- Sustained FPS ≥ 15
- Average latency < 120 ms
- Reasonable CPU usage for 1080p -> terminal 120x40
- Perceptible quality improvement vs basic ASCII

