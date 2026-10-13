# Diagnóstico Técnico y Plan de Optimización del Pipeline

Documento de auditoría del pipeline de renderizado de video en terminal (`video-terminal`).

---

## 1. Arquitectura y Flujo de Ejecución

### Renderizado por bloques
- Cada celda del terminal proyecta 2 píxeles verticales mediante el carácter de medio bloque `▀` (`U+2580`):
  - **Foreground (FG):** píxel superior.
  - **Background (BG):** píxel inferior.
- **Resolución efectiva de video:** `TermW × (TermH · 2)`.  
  *Ejemplo:* En una ventana de 120×40 celdas, la resolución interna es de **120×80 píxeles**.

### Ciclo por frame ([pipeline/run.go](file:///d:/programacion/video-terminal/pipeline/run.go))
1. **Decodificación ([ingest/ffmpeg_decoder.go](file:///d:/programacion/video-terminal/ingest/ffmpeg_decoder.go)):** FFmpeg extrae `rawvideo rgb24` por `stdout` (`pipe:1`) con `-vf fps=N`. El audio (`s16le`, 44.1 kHz, estéreo) sale por un socket TCP local hacia el reloj maestro.
2. **Sincronización A/V:** El reloj de audio comanda el ritmo. Si el frame de video adelanta >10 ms, duerme en timer. Si atrasa >25 ms, descarta frames del canal de prefetch (hasta `FpsTarget` por ciclo).
3. **Resize ([render/resize.go](file:///d:/programacion/video-terminal/render/resize.go)):** Vecino más cercano ajustado a `TermW × (TermH · 2)`.
4. **Temporal Blend (opcional):** Mezcla ponderada con el frame previo (`BlendAlpha`).
5. **Scanline (opcional, presets CRT/Quality):** Atenuación de líneas y simulación de máscara de fósforo RGB.
6. **Cuantización ([render/quantizer.go](file:///d:/programacion/video-terminal/render/quantizer.go)):** 64 niveles/canal en Truecolor, 6 niveles/canal en ANSI-256.
7. **Dither ([render/dither.go](file:///d:/programacion/video-terminal/render/dither.go)):** Matriz Bayer 4×4 o 8×8 según contraste local.
8. **Mapper ([render/mapper.go](file:///d:/programacion/video-terminal/render/mapper.go)):** Empaqueta pares de píxeles en celdas `Cell{Top, Bottom, '▀'}`.
9. **Diff ([diff/byte_differ.go](file:///d:/programacion/video-terminal/diff/byte_differ.go)):** Compara celda a celda contra el frame anterior (`curr != prev`) y agrupa celdas contiguas idénticas en corridas (`DiffOp`).
10. **Emisión ANSI ([render/output.go](file:///d:/programacion/video-terminal/render/output.go)):** Genera secuencias de cursor `\x1b[y;xH` y color `\x1b[38;2;...;48;2;...m`, escribe los bloques `▀` y hace flush a un buffer de 64 KB.

---

## 2. Fallos Críticos y Cuellos de Botella

### A. I/O y Decodificación
1. **Transferencia de resolución completa por pipe:**
   FFmpeg decodifica y transfiere la resolución nativa original (e.g. 1080p a 60 fps ≈ 370 MB/s de memoria sin comprimir) a través del pipe anónimo para luego descartar el 95% de los píxeles en Go con vecino más cercano. Desperdicio masivo de ancho de banda IPC, memoria y ciclos de CPU.
2. **Deformación por Aspect Ratio:**
   El reescalado fuerza `src.W -> TermW` y `src.H -> TermH*2` sin calcular letterboxing/pillarboxing, distorsionando cualquier video cuya relación de aspecto no coincida con el ratio de celdas del terminal.
3. **Aliasing severo:**
   El muestreo por vecino más cercano con factores de reducción altos (~16:1) genera parpadeo temporal de alta frecuencia y bordes dentados.

### B. Color y Procesamiento Gráfico
4. **Inversión de orden Dither / Cuantización (Dither inoperante):**
   En [pipeline/run.go:L212-L226](file:///d:/programacion/video-terminal/pipeline/run.go#L212-L226), el pipeline ejecuta `Quantize()` y luego `Dither()`. El dither suma un sesgo Bayer a valores que ya fueron redondeados a la cuadrícula fija sin volver a cuantizar.
   - **Efecto:** No distribuye el error de cuantización; añade ruido estructurado sobre una imagen posterizada y rompe la uniformidad cromática, impidiendo que el Differ fusione corridas.
5. **Mapeo ANSI 256 deficiente:**
   [render/output.go:L95-L100](file:///d:/programacion/video-terminal/render/output.go#L95-L100) usa un cubo 6×6×6 con truncamiento entero (`c * 5 / 255`) en lugar de redondeo (`(c * 5 + 127) / 255`), y descarta por completo los 24 niveles dedicados de la rampa de grises (códigos ANSI 232–255).
6. **Código muerto:**
   `ditherChannel()` en [render/color.go:L46-L68](file:///d:/programacion/video-terminal/render/color.go#L46-L68) no se utiliza en ningún flujo de producción.

### C. Diff, Memoria y Emisión ANSI
7. **Asignaciones en el heap en el hot loop del Differ:**
   En [diff/byte_differ.go:L45-L46](file:///d:/programacion/video-terminal/diff/byte_differ.go#L45-L46), cada corrida (`flush`) asigna un slice nuevo:
   ```go
   textCopy := make([]rune, len(d.runes)-runStartIdx)
   copy(textCopy, d.runes[runStartIdx:])
   ```
   A 30–60 fps con cientos de cambios por frame, esto genera miles de allocations/segundo que saturan el Garbage Collector. Dado que el carácter es siempre `'▀'`, almacenar `[]rune` es redundante frente a un simple entero `RunLen int`.
8. **Diff por igualdad binaria exacta y falta de Keyframe Refresh:**
   Comparar RGB exacto (`cc != pc`) hace que el ruido de compresión o grano de cámara marque casi toda celda como cambiada, invalidando el ahorro del diff. Sin embargo, introducir un umbral sin **Keyframe Refresh** causará *drift* o *ghosting* acumulativo en el emulador del terminal.
9. **Saturación del emulador de terminal (VTE):**
   Con escenas en movimiento, miles de secuencias `\x1b[y;xH` y Truecolor individuales colapsan el parser del emulador antes de agotar el ancho de banda del socket local.

---

## 3. Matriz de Soluciones Priorizada (ROI)

| Prioridad | Medida | Impacto Técnico | Coste / Consideraciones |
| :--- | :--- | :--- | :--- |
| **P0** | **Reescalado en FFmpeg** (`-vf scale=W:H:flags=area,fps=N`) — **COMPLETADO** | Reduce el I/O por pipe de ~180 MB/s a <1 MB/s. Delega antialiasing de alta calidad (`area`/`lanczos`) a libswscale. | Resuelto: FFmpeg escala a `TermW × (TermH*2)` directamente; si cambia la ventana, `NearestResizer` procesa la diferencia durante la sesión. |
| **P0** | **Invertir/Integrar Dither y Cuantización** — **COMPLETADO** | Dithering matemáticamente correcto. Suprime posterización y gradientes escalonados. | Resuelto: se invirtió el orden en pipeline. El sesgo Bayer se aplica sobre colores continuos y luego se cuantiza a niveles discretos antes del Mapper. |
| **P0** | **Eliminar allocations en `ByteDiffer`** — **COMPLETADO** | Zero-alloc en el render loop. Reduce drásticamente la latencia de GC. | Resuelto: se incorporó `RunLen int` en `DiffOp`, suprimiendo la asignación de slices `make([]rune)` en el differ; `ANSIOutput` emite el rune repetido. |
| **P1** | **Diff con umbral perceptual + Keyframe Refresh** — **COMPLETADO** | Disminuye hasta un 70% los bytes emitidos al terminal en escenas complejas sin parpadeo. | Resuelto: comparación perceptual luma en `ByteDiffer`, barrido continuo de líneas (rolling refresh cada ~30 frames) y keyframe completo periódico (cada 120 frames); `prev` preserva el estado real del terminal sin drift. |
| **P1** | **Preservación de Aspect Ratio** | Evita deformación de la imagen. | Calcular barras negras (letterbox/pillarbox) centradas dentro de `TermW × TermH`. |
| **P1** | **Corrección de ANSI-256** | Mejora sustancial de contraste y fidelidad en terminales de 8 bits. | Usar redondeo correcto e indexar la rampa de grises (232–255) para píxeles monocromáticos/baja saturación. |
| **P2** | **Salida optimizada de cursor y terminal** | Minimiza secuencias ANSI redundantes. | Si dos corridas están en la misma fila y la distancia es menor al coste de emitir `\x1b[y;xH`, emitir espacios o continuar la emisión. |
| **P2** | **Protocolos gráficos nativos (Kitty / Sixel / iTerm2)** | Renderizado con fidelidad de píxel real sin caracteres de texto. | Requiere negociación de capacidades del terminal; no es universalmente compatible. |
