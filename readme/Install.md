# Instalación sin Go (usuario final)

Objetivo: ejecutar `vterminal` sin tener Go instalado.

## Requisitos
- FFmpeg instalado y disponible en el `PATH`.

## Paso 1 — Descargar binario
Descarga el binario correcto para tu sistema operativo y arquitectura.

Ejemplos de nombres:
- `vterminal_windows_amd64.exe`
- `vterminal_linux_amd64`
- `vterminal_darwin_arm64`

## Paso 2 — Mover a una carpeta en el PATH (por OS)

### Windows
1. Crea una carpeta, por ejemplo `C:\Tools\bin\`
2. Mueve `vterminal_windows_amd64.exe` a esa carpeta
3. Agrega la carpeta al `PATH`:
```powershell
[Environment]::SetEnvironmentVariable("Path", $env:Path + ";C:\Tools\bin", "User")
```
4. Abre una nueva terminal y valida:
```powershell
vterminal --help
```

### macOS
1. Mueve el binario a `/usr/local/bin`:
```bash
sudo mv vterminal_darwin_arm64 /usr/local/bin/vterminal
sudo chmod +x /usr/local/bin/vterminal
```
2. Valida:
```bash
vterminal --help
```

### Linux
1. Mueve el binario a `/usr/local/bin`:
```bash
sudo mv vterminal_linux_amd64 /usr/local/bin/vterminal
sudo chmod +x /usr/local/bin/vterminal
```
2. Valida:
```bash
vterminal --help
```

## Paso 3 — Verificar instalación
```bash
vterminal --help
```

## Paso 4 — Ejecutar
```bash
vterminal --input .\test.mp4
```

Si FFmpeg no está en el PATH, indica rutas explícitas:
```bash
vterminal --input .\test.mp4 --ffmpeg C:\ffmpeg\bin\ffmpeg.exe --ffprobe C:\ffmpeg\bin\ffprobe.exe
```
