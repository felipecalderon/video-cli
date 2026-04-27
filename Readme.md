# VTerminal — real-time video rendering engine for the terminal (with audio sync)

![demo GIF](/public/vterminal-presentation.gif)

Real-time video rendering in the terminal using ANSI truecolor, perceptual dithering and synchronized audio.

## Install

- Download from GitHub Releases
- Or build from source:

```powershell
# Windows
.\scripts\build.ps1
```

```bash
# macOS / Linux
chmod +x ./scripts/build.sh && ./scripts/build.sh
```

After install, run the CLI as `vterminal`.

## Quick start

Play a local file:

```bash
vterminal --input ./test.mp4
```

Play a YouTube video (requires [yt-dlp](https://github.com/yt-dlp/yt-dlp) installed):

```bash
vterminal --input https://youtube.com/...
```

If you prefer to run from source:

```bash
go run ./cmd/vterminal --input ./test.mp4
```

## Highlights

1. Audio + video synchronization (primary differentiator)
2. Real-time rendering performance
3. Diff-based incremental rendering (minimal stdout)
4. Perceptual optimizations: channel dithering, subpixel simulation
5. Truecolor (24-bit) when available

## How it works (high level)

Video input (FFmpeg) → Frame buffer → RGB channel split → Adaptive quantization & dithering → Char mapping + ANSI → Diff engine → Terminal output

For design decisions and architecture details see [readme/ARCHITECTURE.md](readme/ARCHITECTURE.md).

## Configuration

See [readme/Configuration.md](readme/Configuration.md) for full configuration options and examples. Short example is available in `config.example.json`.

## Benchmarks

Benchmarks and performance numbers are in progress. See readme/Todo.md for planned benchmark tasks.

## Examples & advanced usage

- Pipe from yt-dlp, use presets, or tune `--fps`/`--preset`
- See [readme/MOCK_FLOW.md](readme/MOCK_FLOW.md) for common pipelines

## Contributing

Bug reports and PRs welcome. See [CONTRIBUTING.md](CONTRIBUTING.md).

## License

MIT
