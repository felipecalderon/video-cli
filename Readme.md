# Motor Perceptual Adaptativo para Terminal

## Descripción

**Motor perceptual adaptativo** es una librería diseñada para renderizar contenido visual (imágenes/video) en terminales, maximizando la calidad percibida mediante técnicas inspiradas en sistemas de visualización antiguos (CRT) y limitaciones modernas de terminal.

En lugar de representar píxeles reales, el motor:

- Optimiza para la **percepción humana**
- Utiliza **mezcla RGB separada**
- Aplica **dithering adaptativo por canal**
- Simula **subpíxeles y persistencia temporal**

---

## Objetivo del MVP (1 mes)

Lograr reproducir un video corto en terminal con:

- ≥ 15 FPS estables
- Uso de color ANSI (idealmente 24-bit)
- Mejora perceptual frente a render tradicional (ASCII simple)
- Implementación de al menos:
  - dithering por canal RGB
  - uso de caracteres Unicode parciales (`▀`, `▄`)
  - diff rendering (no redibujar todo)

---

## Estado actual

- Semana 1: base funcional completa
- Semana 2: motor perceptual inicial completo
- Semana 3: diff rendering completo, buffer doble y pacing en progreso
- Semana 4: dithering dinámico optimizado, scanlines y modo CRT ya integrados

---

## Características clave

- Render perceptual (no pixel-perfect)
- Separación de canales RGB
- Subdivisión de celdas con Unicode
- Dithering espacial adaptativo
- Persistencia temporal (simulación CRT)
- Render incremental (diff-based)
- Escalado dinámico según tamaño de terminal

---

## Arquitectura (alto nivel)

```
Video Input (FFmpeg)
        ↓
Frame Buffer (RGB)
        ↓
Separación de canales (R, G, B)
        ↓
Cuantización adaptativa
        ↓
Dithering por canal
        ↓
Mapping a caracteres + ANSI
        ↓
Buffer de render
        ↓
Diff engine
        ↓
Terminal output
```

---

## Stack sugerido

- **Core**: Go (concurrencia + performance)
- **Decodificación**: FFmpeg
- **Opcional**: C/C++ para optimizaciones críticas
- **CLI / tooling**: TypeScript

---

## Ejemplo de uso (conceptual)

```bash
go run ./cmd/player --input .\test.mp4 --fps 15 --color auto --preset fast
```

Si tu FFmpeg no está en el PATH:

```bash
go run ./cmd/player --input .\test.mp4 --fps 15 --color auto --preset fast --ffmpeg C:\ffmpeg\bin\ffmpeg.exe --ffprobe C:\ffmpeg\bin\ffprobe.exe
```

---


## Configuración (JSON)

Puedes usar un archivo de configuración y sobreescribirlo desde CLI.

Ejemplo (`config.json`):

```json
{
  "fps": 15,
  "preset": "quality",
  "color": "truecolor",
  "scale": 1.0,
  "term_width": 160,
  "term_height": 45
}
```

Uso:

```bash
go run ./cmd/player --input .\test.mp4 --config .\config.example.json
```

Campos soportados:
- `fps`
- `preset` (`fast` | `quality` | `crt`)
- `color` (`auto` | `truecolor` | `256`)
- `scale` (multiplica el tamaño detectado del terminal)
- `term_width` (ancho en columnas)
- `term_height` (alto en filas)

Prioridad: CLI sobrescribe configuración. Si no pasas --config, se carga ./config.json si existe.
Validación estricta: campos desconocidos o valores inválidos generan error.

Flags CLI útiles:
- `--config` (ruta a JSON)
- `--fps` (FPS objetivo)
- `--preset` (`fast` | `quality` | `crt`)
- `--color` (`auto` | `truecolor` | `256`)
- `--scale` (multiplica tamaño detectado)
- `--term-width` (ancho en columnas)
- `--term-height` (alto en filas)
## Limitaciones conocidas

- Dependencia del rendimiento de stdout
- Variabilidad entre terminales
- Sin acceso real a subpíxeles
- Sensibilidad a patrones mal calibrados (ruido visual)

---









