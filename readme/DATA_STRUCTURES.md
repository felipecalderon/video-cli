# Data Structures — Initial Draft

## Basic units (decide explicitly)

- **Source pixel**: from the original frame (RGB 8-bit)
- **Work pixel**: post-resize (RGB 8-bit or float)
- **Terminal cell**: 1 character (2 vertical subcells using `▀`/`▄`)
- **Subcell**: top/bottom half of a cell

---

## Minimum required buffers

### 1) FrameBuffer (source)
- Dimensions: `srcW`, `srcH`
- Format: `RGB24` (3 bytes per pixel)
- Layout: interleaved `RGBRGB...` (MVP decision)

### 2) WorkBuffer (resized)
- Dimensions: `termW`, `termH*2` (if `▀`/`▄` are used)
- Format: `RGB24` sRGB (MVP decision, no linearization)
- Purpose: base for dithering and subcell mapping

### 3) ChannelBuffers (R/G/B)
- 3 buffers of `uint8` or `float`
- Optional: views/slices into WorkBuffer to avoid copies

### 4) DitherState
- Bayer matrix or error-diffusion state
- Persistent state (if temporal dithering is used)

### 5) CellBuffer (render target)
- Dimensions: `termW x termH`
- Each cell contains:
  - `rTop,gTop,bTop`
  - `rBottom,gBottom,bBottom`
  - `rFinal,gFinal,bFinal` (if blending applied)
  - `char` (`▀`, `▄`, ` `)

### 6) DiffBuffer (previous frame)
- Per-cell hash or direct comparison data
- Optional: store last emitted ANSI sequences

---

## Formats and precision (decide)

- Decision: operate in `sRGB` directly (no linearization) for MVP
- Decision: quantize to 6 bits per channel for truecolor; 6x6x6 palette for 256-color mode
- Question: apply spatial or temporal dithering first?

---

## Temporal persistence model (if applicable)

- Accumulate N frames (simple EMA)
- Blend: `new = alpha*current + (1-alpha)*prev`
- Requires a `TemporalBuffer` in float

---

## Recommended structures (Go draft)

```go
type Frame struct {
  W, H int
  Pix  []uint8 // RGBRGB...
  Stride int
}

type Cell struct {
  Top  [3]uint8
  Bottom [3]uint8
  Ch rune
}

type CellGrid struct {
  W, H int
  Cells []Cell
}

type DiffGrid struct {
  W, H int
  Hash []uint32
}
```
