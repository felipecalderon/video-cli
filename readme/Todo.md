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

## Pendientes (Opcional)
- [x] Comando global `vterminal` (entrypoint `cmd/vterminal`)
- [x] Scripts de build cross-platform (Windows/macOS/Linux)
- [ ] Empaquetado con binarios precompilados
- [ ] Publicar releases con binarios
- [ ] Guía de instalación sin Go (FFmpeg externo)
- [ ] Cachear contraste por tile en dithering dinámico
- [ ] Agrupar ops contiguos por fila para menos movimientos de cursor
- [ ] Resize optimizado (SIMD / Go assembly si se necesita)
- [ ] Resize dinámico al cambiar tamaño de terminal
