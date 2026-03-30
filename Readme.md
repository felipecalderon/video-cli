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

- MVP funcional y optimizado (render fluido, sin lag perceptible)
- Pendiente: empaquetado con binarios y CLI global `vterminal`

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

## Ejemplo de uso (actual)

```bash
go run ./cmd/vterminal --input .\test.mp4 --fps 15 --color auto --preset fast
```

Si tu FFmpeg no está en el PATH:

```bash
go run ./cmd/vterminal --input .\test.mp4 --fps 15 --color auto --preset fast --ffmpeg C:\ffmpeg\bin\ffmpeg.exe --ffprobe C:\ffmpeg\bin\ffprobe.exe
```

---

## Dependencias
- FFmpeg
- yt-dlp

## Build cross-platform

### Windows (PowerShell)
```powershell
.\scripts\build.ps1
```

### macOS / Linux (bash)
```bash
chmod +x ./scripts/build.sh # Dar permisos de ejecución.
```

```bash
./scripts/build.sh
```

Los binarios quedan en `dist/` con nombres como `vterminal_windows_amd64.exe`.

---

## Instalación sin Go

Ver guía: `readme/Install.md`.

## Instalación (futuro, sin Go)

Objetivo: que el usuario ejecute solo:

```bash
vterminal --input .\test.mp4
```

### Plan de empaquetado (binarios precompilados)
1. Definir comando global: `vterminal`
2. Mantener `cmd/vterminal` como entrypoint único
3. Generar binarios para Windows/macOS/Linux con `go build`
4. Publicar releases con los binarios
5. Documentar instalación: descarga binario, agregar al `PATH` y requerir FFmpeg externo

Nota: FFmpeg seguirá siendo requisito externo.


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
go run ./cmd/vterminal --input .\test.mp4 --config .\config.example.json
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









