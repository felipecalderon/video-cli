# Critical Open Questions (before implementing)
# Status: RESOLVED (MVP)

## Terminal and output
- Is 24-bit ANSI assumed? Fallback to 256?
  - Decision: default to `truecolor`. Fallback to 256-color if `COLORTERM` is not `truecolor` and `TERM` does not contain `256color`.
  - CLI: `--color=auto|truecolor|256` to force mode.
- Detect `TERM` + capabilities or assume heuristics?
  - Decision: simple heuristics for MVP (env vars). Full terminfo is postponed.
- How to handle live resize?
  - Decision: listen to `SIGWINCH`, recompute `termW/termH`, reallocate buffers, drop current frame and continue.

## Video input
- `ffmpeg` as an external process or embedded library?
  - Decision: `ffmpeg` as external process (pipe `rawvideo rgb24` over stdout).
- Is audio supported? (probably not for MVP)
  - Decision: no audio in MVP.
- Fixed FPS vs video timestamps?
  - Decision: fixed target FPS (e.g., 15), enforced in `ffmpeg` via `-vf fps=`. Ignore real timestamps.

## Performance
- CPU/RAM budget target?
  - Decision: MVP target <= 1 sustained core, RAM < 200 MB.
- What terminal size is considered “normal”?
  - Decision: 120x40 (cols x rows).
- Prioritize FPS or perceptual quality?
  - Decision: stable FPS first, perceptual quality second.

## Dithering and perception
- Fixed Bayer or region-adaptive?
  - Decision: fixed Bayer (4x4 and 8x8), no adaptive dithering in MVP.
- Are there quality presets?
  - Decision: `fast` (4x4), `quality` (8x8), `crt` (8x8 + temporal).
- How to measure "perceptual improvement"?
  - Decision: visual evaluation + MSE comparison vs ASCII baseline (for internal tuning only).

## Diff rendering
- Hash per cell or direct byte comparison?
  - Decision: direct byte comparison of `Cell` (top/bottom RGB + char).
- Is flicker considered a "change"?
  - Decision: no. Only real buffer changes.

## Color format
- Linearize or stay in sRGB?
  - Decision: sRGB directly for MVP (no linearization).
- Quantization per channel (bits per channel)?
  - Decision: 6 bits per channel (truecolor). For 256-color: 6x6x6 palette (216).
