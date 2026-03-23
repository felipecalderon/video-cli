# Preguntas abiertas críticas (antes de programar)
# Estado: RESUELTAS (MVP)

## Terminal y output
- ¿Se asume soporte 24-bit ANSI? ¿Fallback 256?
  - Decision: `truecolor` por defecto. Fallback a 256 si `COLORTERM` no es `truecolor` y `TERM` no contiene `256color`.
  - CLI: `--color=auto|truecolor|256` para forzar.
- ¿Se detecta `TERM` + capabilities o se asume fijo?
  - Decision: heuristica simple en MVP (env vars). Terminfo completo se posterga.
- ¿Cómo se maneja resize en tiempo real?
  - Decision: escuchar `SIGWINCH`, recomputar `termW/termH`, realocar buffers, descartar frame en curso y continuar.

## Video input
- ¿`ffmpeg` como proceso externo o librería embebida?
  - Decision: `ffmpeg` como proceso externo (pipe `rawvideo rgb24` por stdout).
- ¿Se soporta audio? (probablemente no para MVP)
  - Decision: no audio en MVP.
- ¿FPS fijo vs timestamp real del video?
  - Decision: FPS fijo objetivo (ej. 15), forzado en `ffmpeg` con `-vf fps=`. Ignorar timestamps reales.

## Rendimiento
- ¿Presupuesto de CPU/RAM objetivo?
  - Decision: objetivo MVP <= 1 core sostenido, RAM < 200 MB.
- ¿Qué tamaño de terminal se considera “normal”?
  - Decision: 120x40 (cols x rows).
- ¿Se prioriza FPS o calidad perceptual?
  - Decision: FPS estable primero, calidad perceptual segunda.

## Dithering y percepción
- ¿Bayer fijo o adaptativo por región?
  - Decision: Bayer fijo (4x4 y 8x8), sin adaptativo en MVP.
- ¿Hay presets de calidad?
  - Decision: `fast` (4x4), `quality` (8x8), `crt` (8x8 + temporal).
- ¿Cómo se mide “mejora perceptual”?
  - Decision: evaluacion visual + comparacion de error MSE vs baseline ASCII simple (solo para tuning interno).

## Diff rendering
- ¿Hash por celda o comparar bytes directos?
  - Decision: comparar bytes directos de `Cell` (top/bottom RGB + char).
- ¿Se considera parpadeo como “cambio”?
  - Decision: no. Solo cambios reales en el buffer.

## Formato de color
- ¿Linearizar o quedarse en sRGB?
  - Decision: sRGB directo en MVP (sin linearizar).
- ¿Cuantización por canal (bits por canal)?
  - Decision: 6 bits por canal (truecolor). En 256-color: paleta 6x6x6 (216).
