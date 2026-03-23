package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"video-terminal/diff"
	"video-terminal/ingest"
	"video-terminal/pipeline"
	"video-terminal/render"
	"video-terminal/term"
	"video-terminal/types"
)

func main() {
	input := flag.String("input", "", "Path to video file")
	fps := flag.Int("fps", 15, "Target FPS")
	color := flag.String("color", "auto", "Color mode: auto|truecolor|256")
	preset := flag.String("preset", "fast", "Preset: fast|quality|crt")
	ffmpegPath := flag.String("ffmpeg", "", "Path to ffmpeg binary")
	ffprobePath := flag.String("ffprobe", "", "Path to ffprobe binary")
	flag.Parse()

	if strings.TrimSpace(*input) == "" {
		slog.Error("missing required --input")
		os.Exit(2)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	resolvedFFmpeg, err := resolveBinaryPath(*ffmpegPath, "ffmpeg")
	if err != nil {
		printBinaryHelp("ffmpeg", err, *ffmpegPath)
		os.Exit(1)
	}

	resolvedFFprobe, err := resolveBinaryPath(*ffprobePath, "ffprobe")
	if err != nil {
		printBinaryHelp("ffprobe", err, *ffprobePath)
		os.Exit(1)
	}

	decoder, err := ingest.NewFFmpegDecoder(ctx, *input, *fps, resolvedFFmpeg, resolvedFFprobe)
	if err != nil {
		slog.Error("failed to initialize decoder", "err", err)
		os.Exit(1)
	}
	defer func() {
		if err := decoder.Close(); err != nil && !errors.Is(err, io.EOF) {
			slog.Warn("decoder close error", "err", err)
		}
	}()

	termW, termH := term.GetSize()
	mode := term.ResolveColorMode(*color)

	fmt.Print("\x1b[2J\x1b[H\x1b[?25l")
	defer fmt.Print("\x1b[0m\x1b[?25h\n")

	p := pipeline.Pipeline{
		Decoder:   decoder,
		Resizer:   render.NearestResizer{},
		Quantizer: render.ChannelQuantizer{},
		Dither:    render.BayerDither{},
		Mapper:    render.BlockMapper{},
		Differ:    diff.ByteDiffer{},
		Output:    render.NewANSIOutput(os.Stdout, mode),
	}

	params := types.PipelineParams{
		TermW:     termW,
		TermH:     termH,
		FpsTarget: *fps,
		ColorMode: mode,
		Preset:    parsePreset(*preset),
	}

	if err := p.Run(ctx, params); err != nil && !errors.Is(err, context.Canceled) {
		slog.Error("pipeline failed", "err", err)
		os.Exit(1)
	}
}

func resolveBinaryPath(explicitPath, name string) (string, error) {
	if strings.TrimSpace(explicitPath) != "" {
		if _, err := os.Stat(explicitPath); err != nil {
			return "", fmt.Errorf("%s path is invalid: %s", name, explicitPath)
		}
		return explicitPath, nil
	}

	path, err := exec.LookPath(name)
	if err != nil {
		return "", fmt.Errorf("%s not found in PATH; install ffmpeg or pass --%s=/full/path/to/%s", name, name, name)
	}

	return path, nil
}

func printBinaryHelp(name string, err error, explicitPath string) {
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintf(os.Stderr, "No pude encontrar %s.\n", name)
	fmt.Fprintf(os.Stderr, "Detalle: %v\n", err)
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "Opciones:")
	fmt.Fprintf(os.Stderr, "  1. Instala FFmpeg y agrega su carpeta bin al PATH.\n")
	fmt.Fprintf(os.Stderr, "  2. Pasa la ruta completa con --%s.\n", name)
	fmt.Fprintf(os.Stderr, "  3. Verifica que el archivo exista en la ruta indicada.\n")
	fmt.Fprintln(os.Stderr, "")
	if strings.TrimSpace(explicitPath) != "" {
		fmt.Fprintf(os.Stderr, "Ruta indicada: %s\n", filepath.Clean(explicitPath))
	}
	fmt.Fprintln(os.Stderr, "Ejemplo:")
	fmt.Fprintf(os.Stderr, "  go run ./cmd/player --input .\\test.mp4 --%s C:\\ffmpeg\\bin\\%s.exe\n", name, name)
}

func parsePreset(v string) types.Preset {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "quality":
		return types.PresetQuality
	case "crt":
		return types.PresetCRT
	default:
		return types.PresetFast
	}
}
