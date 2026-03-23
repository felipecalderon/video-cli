# Estructuras de datos — Borrador Inicial

## Unidades básicas (decidir explícitamente)

- **Pixel fuente**: del frame original (RGB 8-bit)
- **Pixel de trabajo**: post-resize (RGB 8-bit o float)
- **Celda terminal**: 1 carácter (2 subceldas verticales con `▀`/`▄`)
- **Subcelda**: mitad superior/inferior de una celda

---

## Buffers mínimos necesarios

### 1) FrameBuffer (source)
- Dimensiones: `srcW`, `srcH`
- Formato: `RGB24` (3 bytes por pixel)
- Layout: interleaved `RGBRGB...` (decision MVP)

### 2) WorkBuffer (resized)
- Dimensiones: `termW`, `termH*2` (si usamos `▀`/`▄`)
- Formato: `RGB24` sRGB (decision MVP, sin linearizar)
- Uso: base para dithering y mapeo a subceldas

### 3) ChannelBuffers (R/G/B)
- 3 buffers `uint8` o `float`
- Opcional: compartir con `WorkBuffer` (vista / slice)

### 4) DitherState
- Matriz Bayer o error diffusion
- Estado persistente (si hay dithering temporal)

### 5) CellBuffer (render target)
- Dimensiones: `termW x termH`
- Cada celda contiene:
  - `rTop,gTop,bTop`
  - `rBottom,gBottom,bBottom`
  - `rFinal,gFinal,bFinal` (si hay fusión)
  - `char` (`▀`, `▄`, ` `)

### 6) DiffBuffer (prev frame)
- Hash por celda o comparación directa
- Opcional: también guardar último ANSI emitido

---

## Formatos y precisión (decidir)

- Decision: trabajar en `sRGB` directo (sin linearizar) para MVP
- Decision: cuantizacion 6 bits por canal (truecolor); 6x6x6 en 256-color
- ¿Dithering spatial o temporal primero?

---

## Modelo de persistencia temporal (si aplica)

- Acumular N frames (simple EMA)
- Mezcla: `new = alpha*current + (1-alpha)*prev`
- Necesita `TemporalBuffer` en float

---

## Estructuras recomendadas (borrador Go)

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
