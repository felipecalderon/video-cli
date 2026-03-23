# Contrato minimo del pipeline (MVP)

Objetivo: definir etapas y tipos para permitir implementar en paralelo sin ambiguedades.

## Tipos base

```go
// RGB interleaved, sRGB, 8-bit
type FrameRGB struct {
  W, H  int
  Stride int       // bytes per row, >= W*3
  Pix   []uint8    // len >= Stride*H
}

// Work buffer ya escalado a terminal, con subceldas verticales
// H = termH*2
type WorkRGB struct {
  W, H  int
  Stride int
  Pix   []uint8
}

type Cell struct {
  Top    [3]uint8  // RGB
  Bottom [3]uint8  // RGB
  Ch     rune      // '▀' o '▄' o ' '
}

type CellGrid struct {
  W, H  int
  Cells []Cell // len = W*H
}
```

## Etapas (I/O)

1. Decode
   - In: bytes `rawvideo rgb24`
   - Out: `FrameRGB` (W=srcW, H=srcH)

2. ResizeToTerm
   - In: `FrameRGB`
   - Out: `WorkRGB` con `W=termW`, `H=termH*2`

3. Quantize
   - In: `WorkRGB` (sRGB 8-bit)
   - Out: `WorkRGB` con cuantizacion 6bpc (modo truecolor)
   - Nota: en modo 256 se cuantiza a 6x6x6 para mapping a paleta.

4. DitherBayer
   - In: `WorkRGB`
   - Out: `WorkRGB` dithered (Bayer 4x4 o 8x8 segun preset)

5. MapCells
   - In: `WorkRGB`
   - Out: `CellGrid` (usa pares verticales -> `▀`/`▄`)

6. Diff
   - In: `CellGrid` + `CellGrid` prev
   - Out: lista de comandos ANSI (solo celdas cambiadas)
   - Criterio: comparar bytes directos de `Cell`

7. Output
   - In: comandos ANSI + buffer bytes
   - Out: stdout

## Parametros minimos globales

- `termW`, `termH` (dinamico, por SIGWINCH)
- `fpsTarget` (ej. 15)
- `colorMode` = auto | truecolor | 256
- `preset` = fast | quality | crt

