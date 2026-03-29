# ✅ TODO LIST — MVP

## Features (Done)

- [x] Integración con FFmpeg (extraer frames)
- [x] Render básico en terminal (color ANSI)
- [x] Resize de frames a tamaño de terminal
- [x] Loop de reproducción simple
- [x] Separación de canales RGB
- [x] Cuantización básica por canal
- [x] Implementación de dithering simple (Bayer)
- [x] Soporte para caracteres Unicode (`▀`, `▄`)
- [x] Persistencia temporal (frame blending)
- [x] Simulación de scanlines
- [x] Modo “CRT” (intensidades y look perceptual)

## Optimización (Done)

- [x] Diff rendering (solo cambios)
- [x] Buffer doble (double buffering)
- [x] Ajuste de FPS estable
- [x] Optimización de escritura a stdout
- [x] Reuso de buffers por frame (evita GC)
- [x] Eliminar goroutine por frame en decoder
- [x] Reuso de timer en pacing de FPS
- [x] Evitar SGR redundante en salida ANSI
- [x] Scanlines sin float por píxel (fixed-point)

## Pendientes

- [x] Comando global `vterminal` (entrypoint `cmd/vterminal`)
- [x] Scripts de build cross-platform (Windows/macOS/Linux)
- [x] Arreglar fuga de memoria de rune a string en Differ
- [ ] Escuchar SIGWINCH para redimensionamiento de terminal en caliente
- [ ] Módulo de frame pacing suave (recuperación o skip de frames atrasados)
- [ ] Corregir el redimensionamiento de terminal que rompe el video, más pequeño se rompe, al agrandar se queda del mismo tamaño original.
