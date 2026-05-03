## Evaluación inicial

El pipeline actual (`pipeline/run.go`) es una cadena secuencial en un solo goroutine: cada frame bloquea todas las etapas antes de pasar al siguiente. Esto tiene dos consecuencias directas:

1. No hay manera de saber cuál etapa es el cuello de botella — perfilar sin instrumentación es adivinanza.
2. Cualquier intento de añadir concurrencia antes de medir va a optimizar la etapa equivocada.

El orden correcto es: **medir primero, concurrenciar después**. Lo que sigue respeta ese orden.

---

## Fase 1 — Infraestructura de profiling

### 1.1 Ring buffer de timings por etapa

Definir una estructura `StageTimer` con un ring buffer de duración fija (ej. últimos 120 frames). Cada etapa del pipeline recibe un puntero a su timer. La medición es un `time.Now()` antes y un `Mark(duration)` después — sin allocations, sin mutex en el hot path (puede usar un índice atómico).

Las etapas a instrumentar son exactamente las del pipeline actual:

- `Decode` (incluye el `ReadFull` sobre el pipe de FFmpeg)
- `Resize`
- `TemporalBlend`
- `Scanline`
- `Quantize`
- `Dither`
- `Map`
- `Diff`
- `Output.Write + Flush`
- `SyncWait` (el tiempo que el goroutine duerme esperando al audio)
- `FrameTotal` (duración de punta a punta por frame)

### 1.2 Métricas derivadas

Del ring buffer calcular en tiempo real: p50, p95, p99 por etapa. Además:

- **Frame budget**: `targetDuration - FrameTotal`. Negativo = el pipeline no cumple el FPS objetivo.
- **AV drift**: `pts - audioClock.CurrentTime()` en el momento de renderizado — cuánto está adelantado o atrasado el video respecto al audio.
- **Skip rate**: frames descartados por segundo por el bloque de sincronía.
- **GC pressure**: `runtime.ReadMemStats` muestreado cada N frames (no por frame), para detectar allocations que escapan a la pila.

### 1.3 Modo de exposición

Agregar un flag `--profile` que activa un overlay en la última fila del terminal. El overlay usa el mismo `ANSIOutput` pero con coordenadas reservadas (última fila). No requiere un goroutine separado — se escribe al final de cada `Output.Write`.

Formato sugerido (una línea, ej.):

```
fps:18/30  decode:2.1ms  dither:8.4ms  output:1.2ms  sync:+12ms  skip:0/s
```

Alternativa para debugging offline: `--profile-log path` que escribe JSON con los percentiles cada segundo a un archivo.

---

## Fase 2 — Protocolo de medición

Antes de tocar el pipeline concurrente, ejecutar estos escenarios con el profiling activado y documentar los resultados:

**Escenario base**: 1080p → terminal 160×45, 30fps objetivo, preset `quality`.
**Escenario degradado**: misma config pero terminal de 220×60 (más carga en dither y output).
**Escenario I/O**: stream remoto via yt-dlp a 480p, 15fps.

Las preguntas que deben quedar respondidas:

1. ¿Qué etapa consume más del presupuesto por frame?
2. ¿El SyncWait es positivo (el video espera al audio) o negativo (el video se atrasa)?
3. ¿El output.Flush es el cuello de botella en terminales lentos?
4. ¿Hay allocations que escapan al heap (ver el GC counter)?

Solo después de esto se toman decisiones de diseño concurrente.

---

## Fase 3 — Rediseño concurrente (condicional a los hallazgos)

Hay tres cuellos de botella posibles y cada uno tiene una solución distinta. No son mutuamente excluyentes, pero conviene implementarlos en orden de impacto.

### 3A — Si el cuello es `Decode` (I/O bound)

El `ReadFull` sobre el pipe de FFmpeg bloquea el goroutine principal mientras espera bytes. La solución es un **prefetch goroutine**:

- Un goroutine dedicado solo a leer frames del decoder y enviarlo a un canal con profundidad 2.
- El goroutine de procesamiento consume del canal.
- Resultado: mientras se procesa el frame N, el decoder ya está leyendo el frame N+1 desde FFmpeg.
- Profundidad del canal: máximo 2 frames (no más, para no adelantar al audio más de `2 / fps` segundos).

Este cambio es el de menor riesgo y no afecta la lógica de sincronía ni de seek.

### 3B — Si el cuello es `Dither` o `Quantize` (CPU bound)

Dividir el pipeline en 3 etapas concurrentes con canales de profundidad 1:

```
Goroutine A: Decode → Resize
Goroutine B: TemporalBlend → Scanline → Quantize → Dither → Map
Goroutine C: Diff → Output
```

La profundidad 1 es intencional: permite que A y B se solapen en tiempo, pero no acumula frames que adelantarían el video al audio.

El punto crítico aquí es la **propiedad del buffer**: cuando el goroutine A pasa un `WorkRGB` al canal hacia B, A no debe tocar ese buffer hasta que B haya terminado. Eso se resuelve con un pool explícito de 3 buffers (ver Fase 4).

### 3C — Si el cuello es `Output.Flush` (I/O bound hacia stdout)

Hacer el write asíncrono: el goroutine de procesamiento genera el slice `[]byte` con las secuencias ANSI y lo envía a un canal. Un goroutine escritor es el único que toca stdout y hace el `Flush`.

Esto desacopla la generación de ANSI del tiempo de flush. El canal debe tener profundidad 1 — si stdout está lleno, el backpressure llega al procesador y este espera, lo cual es el comportamiento correcto (no tiene sentido generar frames que no se pueden mostrar).

### 3D — Seek y resize en pipeline concurrente

Este es el caso más delicado. El enfoque actual (cancelar el contexto y reiniciar el decoder) es correcto pero lento porque obliga a reiniciar FFmpeg. Con un pipeline concurrente hay goroutines bloqueados en canales que hay que drenar.

El diseño recomendado: **mantener la cancelación de contexto** como mecanismo de reset, pero añadir un paso de drenado explícito. Cuando el contexto se cancela:

1. Todos los goroutines del pipeline escuchan `ctx.Done()` y salen.
2. El goroutine que inicia el seek espera a que todos los canales estén vacíos (puede usar un `sync.WaitGroup` o leer hasta EOF del canal antes de reiniciar).
3. El decoder se reinicia con el nuevo offset.

No intentar hacer seek "caliente" sin reiniciar FFmpeg — la complejidad no vale el beneficio para este caso de uso.

---

## Fase 4 — Modelo de propiedad de buffers

Con concurrencia, los buffers no pueden ser reutilizados libremente. El modelo actual ya tiene buffers reutilizables en `NearestResizer`, `BayerDither`, `TemporalBlend`, `ByteDiffer` — pero solo funciona porque hay un solo goroutine.

Con el pipeline por etapas, cada etapa necesita sus propios buffers. El diseño más simple:

- Pre-alocar un **pool de N slots** donde N = profundidad_canal + 1 (ej. 3 slots para canal de profundidad 1 + 1 en vuelo).
- Cada slot contiene: `FrameRGB` + `WorkRGB` (×2 para double-buffer del temporal blend) + `CellGrid`.
- Los slots circulan a través del pipeline: A toma el slot libre, lo llena, lo envía; B lo procesa, lo envía; C lo consume y lo devuelve al pool libre.
- Implementar con un canal de "slots libres" de profundidad N — clásico y correcto.

Esto elimina todas las allocations por frame una vez que el sistema está en régimen estacionario.

---

## Fase 5 — Benchmarks y validación

### 5.1 Benchmark sintético del pipeline completo

Añadir `pipeline_bench_test.go` con un decoder mock que genera frames pre-generados en memoria (no FFmpeg). Mide el throughput de las etapas de procesamiento en aislamiento. Permite comparar antes/después de cualquier cambio sin depender de un archivo de video o de FFmpeg.

### 5.2 Benchmark por etapa aislada

Ya existen `BenchmarkBayerDitherQuality`, `BenchmarkScanlineCRT`, `BenchmarkTemporalBlend`. Agregar:

- `BenchmarkBlockMapper` (MapInto)
- `BenchmarkByteDiffer` (ya existe)
- `BenchmarkANSIOutput` con un `io.Discard` como writer

Ejecutar con `-benchmem` para detectar allocations que escapan al heap.

### 5.3 Criterios de éxito

Antes de cerrar cualquier cambio de concurrencia:

- FPS sostenido ≥ target en el escenario base.
- AV drift < 25ms en régimen estacionario.
- Zero allocations por frame en las etapas de procesamiento (verificado con `-benchmem`).
- Seek funcional y sin goroutine leaks (verificar con `goleak` o `runtime.NumGoroutine` antes/después).

---

## Orden de ejecución recomendado

1. Ring buffer + StageTimer | Sin esto, todo lo demás es adivinanza |
2. Overlay `--profile` | Necesario para observar en tiempo real |
3. Medir escenarios base | Define qué optimizar |
4. Pool de buffers | Prerequisito de cualquier concurrencia |
5. La 3A/B/C que corresponda | Basado en medición, no en suposición |
6. Benchmarks sintéticos | Valida que el cambio tiene impacto real |
7. Seek/resize en pipeline concurrente | Último porque es lo más propenso a race conditions |
