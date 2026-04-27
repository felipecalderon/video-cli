# Configuration

This document describes the main configuration options for vterminal. You can provide a JSON file (e.g. `config.json`) and override values from the CLI. If no `--config` is passed, the CLI will load `./config.json` if present.

Example (config.json):

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

Fields

- `fps` (number): Target frames per second for rendering. The engine will try to keep close to this value but actual FPS depends on terminal and system performance.
- `preset` (string): Rendering preset. Supported: `fast`, `quality`, `crt`.
- `color` (string): Color mode. Supported: `auto`, `truecolor`, `256`.
- `scale` (number): Multiplier for the detected terminal size. Use values <1 to downscale, >1 to upscale.
- `term_width` / `term_height` (integers): Force a specific terminal size in columns/rows. Useful for reproducible outputs.

Behavior

- CLI flags take precedence over values in the config file.
- Unknown fields or invalid values will produce an error (strict validation).
- Keep the config file next to your working directory or pass `--config path/to/config.json`.

Examples

- Use quality preset with 24-bit color:

```bash
vterminal --config ./config.json --preset quality --color truecolor
```

- Run with a custom FPS without editing the config file:

```bash
vterminal --input test.mp4 --fps 20
```

See also: `config.example.json` at the project root and readme/PIPELINE_CONTRACT.md for pipeline-related options.
