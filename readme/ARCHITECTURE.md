# Arquitectura — Borrador Inicial

## Objetivo
Definir módulos, límites, y flujo de datos para evitar decisiones implícitas cuando empiece el código.

---

## Módulos propuestos

- `ingest/` (FFmpeg + entrada de video)
- `pipeline/` (procesamiento por etapas)
- `render/` (mapeo a celdas + ANSI)
- `diff/` (motor de diferencias)
- `term/` (detección de capacidades + tamaño)
- `timing/` (control de FPS + sincronización)
- `profile/` (stats, métricas, perfiles)

---

## Pipeline lógico (etapas)

1. Decode (frames RGB)
2. Resize a resolución de terminal
3. Conversión de color/linealización (si aplica)
4. Separación de canales
5. Cuantización por canal
6. Dithering por canal
7. Mapeo a celdas (Unicode + ANSI)
8. Diff (frame actual vs anterior)
9. Output (stdout)

---

## Concurrencia (decisión pendiente)

- ¿Pipeline por etapas con canales Go?
- ¿Doble buffer con lock-free?
- ¿Backpressure si stdout es lento?

---

## Integración con FFmpeg (decisión pendiente)

- Entrada: `ffmpeg` -> `rawvideo` por stdout
- Formato: `rgb24` / `rgba` / `bgr24`
- Sincronización: ¿timestamp externo o fijo por FPS objetivo?

---

## Criterios de éxito del MVP (medibles)

- FPS sostenido ≥ 15
- Latencia media < 120 ms
- Uso CPU razonable en 1080p -> terminal 120x40
- Diferencia perceptual clara vs ASCII básico

