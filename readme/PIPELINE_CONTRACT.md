# Minimal pipeline contract (MVP)

Goal: define stages and types so parallel implementation has no ambiguities.

## Base types

```go
// RGB interleaved, sRGB, 8-bit
type FrameRGB struct {
  W, H  int
  Stride int       // bytes per row, >= W*3
  Pix   []uint8    // len >= Stride*H
}

// Work buffer already scaled to terminal, with vertical subcells
// H = termH*2
type WorkRGB struct {
  W, H  int
  Stride int
  Pix   []uint8
}

type Cell struct {
  Top    [3]uint8  // RGB
  Bottom [3]uint8  // RGB
  Ch     rune      // '▀' or '▄' or ' '
}

type CellGrid struct {
  W, H  int
  Cells []Cell // len = W*H
}
```

## Stages (I/O)

1. Decode
   - In: bytes `rawvideo rgb24`
   - Out: `FrameRGB` (W=srcW, H=srcH)

2. ResizeToTerm
   - In: `FrameRGB`
   - Out: `WorkRGB` with `W=termW`, `H=termH*2`

3. Quantize
   - In: `WorkRGB` (sRGB 8-bit)
   - Out: `WorkRGB` quantized to 6bpc (truecolor mode)
   - Note: in 256 mode quantize to 6x6x6 for palette mapping.

4. DitherBayer
   - In: `WorkRGB`
   - Out: `WorkRGB` dithered (Bayer 4x4 or 8x8 depending on preset)

5. MapCells
   - In: `WorkRGB`
   - Out: `CellGrid` (uses vertical pairs -> `▀`/`▄`)

6. Diff
   - In: `CellGrid` + previous `CellGrid`
   - Out: list of ANSI commands (only changed cells)
   - Criterion: direct byte comparison of `Cell`

7. Output
   - In: ANSI commands + buffer bytes
   - Out: stdout

## Minimum global parameters

- `termW`, `termH` (dynamic, via SIGWINCH)
- `fpsTarget` (e.g. 15)
- `colorMode` = auto | truecolor | 256
- `preset` = fast | quality | crt

