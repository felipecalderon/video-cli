# ✅ TODO LIST — MVP

## Features (Done)

- [x] FFmpeg integration (extract frames)
- [x] Basic terminal rendering (ANSI color)
- [x] Resize frames to terminal size
- [x] Simple playback loop
- [x] RGB channel separation
- [x] Basic per-channel quantization
- [x] Simple dithering implementation (Bayer)
- [x] Unicode character support (`▀`, `▄`)
- [x] Temporal persistence (frame blending)
- [x] Scanline simulation
- [x] “CRT” mode (intensity and perceptual look)

## Optimization (Done)

- [x] Diff rendering (only changes)
- [x] Double buffering
- [x] Stable FPS tuning
- [x] Optimized stdout writes
- [x] Reuse buffers per frame (avoid GC)
- [x] Remove per-frame goroutine in decoder
- [x] Reuse timer for FPS pacing
- [x] Avoid redundant SGR in ANSI output
- [x] Scanlines without per-pixel floats (fixed-point)

## Pending

- [x] Global command `vterminal` (entrypoint `cmd/vterminal`)
- [x] Cross-platform build scripts (Windows/macOS/Linux)
- [x] Fix rune-to-string memory leak in Differ
- [ ] Listen to SIGWINCH for hot terminal resize
- [ ] Smooth frame pacing module (recovery or skipping behind frames)
- [ ] Fix terminal resize bug that breaks playback when smaller; enlarging leaves previous size unchanged.
