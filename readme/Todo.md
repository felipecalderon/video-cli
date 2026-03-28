# ✅ TODO LIST — MVP (1 mes)

## Semana 1 — Base funcional

- [x] Integración con FFmpeg (extraer frames)
- [x] Render básico en terminal (color ANSI)
- [x] Resize de frames a tamaño de terminal
- [x] Loop de reproducción simple

---

## Semana 2 — Motor perceptual inicial

- [x] Separación de canales RGB
- [x] Cuantización básica por canal
- [x] Implementación de dithering simple (Bayer)
- [x] Soporte para caracteres Unicode (`▀`, `▄`)

---

## Semana 3 — Optimización

- [x] Diff rendering (solo cambios)
- [x] Buffer doble (double buffering)
- [x] Ajuste de FPS estable
- [x] Optimización de escritura a stdout

---

## Semana 4 — Magia perceptual

- [x] Persistencia temporal (frame blending)
- [x] Ajuste dinámico de dithering
  - [x] Separar el análisis de contraste del buffer de salida
  - [x] Afinar rangos de bias por preset (`quality` / `crt`)
  - [x] Agregar pruebas de estabilidad visual y no mutación
- [x] Simulación de scanlines
- [x] Modo “CRT”
  - [x] Ajuste final de intensidades y look perceptual

---

## Stretch goals (si sobra tiempo)

- [ ] Interlaced rendering
- [ ] Configuración dinámica según terminal
- [ ] Perfiles visuales (retro, limpio, glitch)

---
