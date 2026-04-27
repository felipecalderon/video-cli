# Installation without Go (end user)

Goal: run `vterminal` without having Go installed.

## Requirements
- FFmpeg installed and available on the system `PATH`.

## Step 1 — Download the binary
Download the binary that matches your OS and architecture.

Example names:
- `vterminal_windows_amd64.exe`
- `vterminal_linux_amd64`
- `vterminal_darwin_arm64`

## Step 2 — Move to a folder in your PATH (by OS)

### Windows
1. Create a folder, e.g. `C:\Tools\bin\`
2. Move `vterminal_windows_amd64.exe` to that folder
3. Add the folder to your `PATH`:
```powershell
[Environment]::SetEnvironmentVariable("Path", $env:Path + ";C:\Tools\bin", "User")
```
4. Open a new terminal and verify:
```powershell
vterminal --help
```

### macOS
1. Move the binary to `/usr/local/bin`:
```bash
sudo mv vterminal_darwin_arm64 /usr/local/bin/vterminal
sudo chmod +x /usr/local/bin/vterminal
```
2. Verify:
```bash
vterminal --help
```

### Linux
1. Move the binary to `/usr/local/bin`:
```bash
sudo mv vterminal_linux_amd64 /usr/local/bin/vterminal
sudo chmod +x /usr/local/bin/vterminal
```
2. Verify:
```bash
vterminal --help
```

## Step 3 — Check installation
```bash
vterminal --help
```

## Step 4 — Run
```bash
vterminal --input .\test.mp4
```

If FFmpeg is not on the PATH, pass explicit paths:
```bash
vterminal --input .\test.mp4 --ffmpeg C:\ffmpeg\bin\ffmpeg.exe --ffprobe C:\ffmpeg\bin\ffprobe.exe
```
