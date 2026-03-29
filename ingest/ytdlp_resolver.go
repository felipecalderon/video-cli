package ingest

import (
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// YtdlpResolver implementa StreamResolver usando yt-dlp.
//
// yt-dlp es una herramienta CLI que extrae URLs directas de streams
// desde cientos de plataformas (YouTube, Twitch, Vimeo, etc.).
// Esta implementación se puede reemplazar por cualquier otra que
// cumpla la interfaz StreamResolver.
type YtdlpResolver struct {
	BinaryPath string // Ruta al binario yt-dlp (si vacío, se busca en PATH)
}

func (r *YtdlpResolver) Name() string { return "yt-dlp" }

func (r *YtdlpResolver) resolvedBinary() (string, error) {
	if strings.TrimSpace(r.BinaryPath) != "" {
		return r.BinaryPath, nil
	}
	path, err := exec.LookPath("yt-dlp")
	if err != nil {
		return "", fmt.Errorf("yt-dlp not found in PATH; install it or pass --ytdlp=/path/to/yt-dlp")
	}
	return path, nil
}

func (r *YtdlpResolver) Resolve(ctx context.Context, rawURL string, maxHeight int) (StreamResult, error) {
	bin, err := r.resolvedBinary()
	if err != nil {
		return StreamResult{}, err
	}

	// Construir el formato de selección de calidad.
	// Si maxHeight > 0, pedimos el mejor stream que no supere esa altura.
	// Esto ahorra ancho de banda y CPU para renderizado en terminal.
	formatSelector := "best"
	if maxHeight > 0 {
		h := strconv.Itoa(maxHeight)
		formatSelector = fmt.Sprintf(
			"best[height<=%s]/bestvideo[height<=%s]+bestaudio/best",
			h, h,
		)
	}

	args := []string{
		"--get-url",
		"--get-title",
		"-f", formatSelector,
		"--no-playlist",
		"--no-warnings",
		rawURL,
	}

	cmd := exec.CommandContext(ctx, bin, args...)
	out, err := cmd.Output()
	if err != nil {
		// Intentar extraer stderr para un mensaje más útil.
		if exitErr, ok := err.(*exec.ExitError); ok && len(exitErr.Stderr) > 0 {
			return StreamResult{}, fmt.Errorf("yt-dlp failed: %s", strings.TrimSpace(string(exitErr.Stderr)))
		}
		return StreamResult{}, fmt.Errorf("yt-dlp failed: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	if len(lines) < 2 {
		return StreamResult{}, fmt.Errorf("yt-dlp returned unexpected output (expected title + URL, got %d lines)", len(lines))
	}

	// yt-dlp con --get-title --get-url imprime:
	// Línea 1: título del video
	// Línea 2: URL directa del stream
	title := strings.TrimSpace(lines[0])
	streamURL := strings.TrimSpace(lines[1])

	if streamURL == "" {
		return StreamResult{}, fmt.Errorf("yt-dlp resolved an empty stream URL")
	}

	return StreamResult{
		URL:   streamURL,
		Title: title,
	}, nil
}
