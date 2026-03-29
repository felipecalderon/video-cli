package ingest

import (
	"context"
	"strings"
)

// StreamResult contiene la URL resuelta y metadatos opcionales del stream.
type StreamResult struct {
	URL   string // URL directa al stream multimedia (.m3u8, .mp4, etc.)
	Title string // Título del video/stream (informativo)
}

// StreamResolver resuelve una URL de plataforma (YouTube, Twitch, etc.)
// a una URL directa que FFmpeg pueda consumir.
//
// Cualquier herramienta externa (yt-dlp, streamlink, etc.) debe implementar
// esta interfaz. El core del sistema nunca depende de una herramienta concreta.
type StreamResolver interface {
	// Resolve toma una URL de página web y devuelve la URL directa al stream.
	// maxHeight indica la resolución máxima deseada (ej. 480) para ahorrar
	// ancho de banda y CPU. Un valor de 0 significa sin límite.
	Resolve(ctx context.Context, rawURL string, maxHeight int) (StreamResult, error)

	// Name devuelve un identificador legible de la herramienta ("yt-dlp", "streamlink", etc.)
	Name() string
}

// IsURL detecta si el input del usuario es una URL remota en lugar de un archivo local.
func IsURL(input string) bool {
	lower := strings.ToLower(strings.TrimSpace(input))
	return strings.HasPrefix(lower, "http://") || strings.HasPrefix(lower, "https://")
}
