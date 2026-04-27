package ingest

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os/exec"
	"strings"
	"time"
	"video-terminal/types"
)

type FFmpegDecoder struct {
	cmd         *exec.Cmd
	stdout      io.ReadCloser
	audioReader *io.PipeReader
	audioWriter *io.PipeWriter
	frameSize   int
	width       int
	height      int
	buf         []byte
}

type ffprobeResponse struct {
	Streams []struct {
		Width  int `json:"width"`
		Height int `json:"height"`
	} `json:"streams"`
}

func NewFFmpegDecoder(ctx context.Context, inputPath string, width, height, fps int, ffmpegPath string, isStream bool, startOffset time.Duration) (*FFmpegDecoder, error) {
	if fps <= 0 {
		fps = 15
	}

	if width <= 0 || height <= 0 {
		return nil, fmt.Errorf("invalid video dimensions")
	}

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, fmt.Errorf("create audio listener: %w", err)
	}
	// FIX: eliminado "defer ln.Close()" del cuerpo principal.
	// El ownership del listener pasa a la goroutine de copia de audio,
	// que lo cierra inmediatamente tras aceptar la conexión (ver abajo).
	// Cerrar aquí además causaba una race condition: el defer del cuerpo
	// corría al retornar NewFFmpegDecoder, justo cuando la goroutine
	// acababa de arrancar, resultando en un doble Close concurrente.

	port := ln.Addr().(*net.TCPAddr).Port

	args := []string{
		"-hide_banner",
		"-loglevel", "error",
	}

	if isStream {
		args = append(args,
			"-reconnect", "1",
			"-reconnect_streamed", "1",
			"-reconnect_delay_max", "2",
		)
	}

	if startOffset > 0 {
		args = append(args, "-ss", fmt.Sprintf("%.3f", startOffset.Seconds()))
	}

	args = append(args,
		"-i", inputPath,
		"-vf", fmt.Sprintf("fps=%d", fps),
		"-f", "rawvideo",
		"-pix_fmt", "rgb24",
		"pipe:1",
		"-f", "s16le",
		"-acodec", "pcm_s16le",
		"-ar", "44100",
		"-ac", "2",
		fmt.Sprintf("tcp://127.0.0.1:%d", port),
	)

	cmd := exec.CommandContext(ctx, ffmpegPath, args...)

	// Guardamos stderr para mostrar el motivo real si FFmpeg no conecta.
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		ln.Close()
		return nil, fmt.Errorf("create ffmpeg stdout pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		ln.Close()
		return nil, fmt.Errorf("start ffmpeg: %w", err)
	}

	// Esperamos la conexión de audio para evitar cerrar el listener demasiado pronto.
	type acceptRes struct {
		conn net.Conn
		err  error
	}
	resChan := make(chan acceptRes, 1)
	go func() {
		conn, err := ln.Accept()
		resChan <- acceptRes{conn: conn, err: err}
	}()

	select {
	case res := <-resChan:
		if res.err != nil {
			_ = cmd.Process.Kill()
			ln.Close()
			return nil, fmt.Errorf("audio connection failed: %w", res.err)
		}

		audioReader, audioWriter := io.Pipe()

		// FIX: la goroutine toma ownership completo del listener y la conexión.
		//
		// Cambios respecto al original:
		//   1. ln.Close() se llama al INICIO de la goroutine, no al final.
		//      El listener solo sirve para aceptar la conexión inicial; una vez
		//      aceptada ya no se necesita y puede cerrarse de inmediato.
		//      Antes se cerraba al final, conviviendo con el defer del cuerpo
		//      principal — dos cierres concurrentes sobre el mismo fd.
		//
		//   2. Se reemplaza io.Copy por un loop manual con select sobre ctx.Done().
		//      io.Copy es bloqueante y no respeta cancelación de contexto: si el
		//      usuario hacía seek o salía, la goroutine quedaba huérfana hasta que
		//      FFmpeg cerraba la conexión por su cuenta (potencialmente nunca en
		//      streams). El loop manual permite salir limpiamente en O(32KB) de latencia.
		go func() {
			// Cerramos el listener aquí: ya aceptamos la única conexión que
			// necesitábamos y no queremos mantener el fd abierto innecesariamente.
			ln.Close()
			defer res.conn.Close()

			buf := make([]byte, 32*1024)
			for {
				// Chequeamos cancelación antes de cada Read para salir
				// sin esperar a que FFmpeg cierre la conexión.
				select {
				case <-ctx.Done():
					_ = audioWriter.CloseWithError(ctx.Err())
					return
				default:
				}

				n, err := res.conn.Read(buf)
				if n > 0 {
					if _, werr := audioWriter.Write(buf[:n]); werr != nil {
						// El lector cerró el pipe (e.g. Close() en el decoder);
						// no hay a dónde escribir, salimos silenciosamente.
						return
					}
				}
				if err != nil {
					if err == io.EOF {
						_ = audioWriter.Close()
					} else {
						_ = audioWriter.CloseWithError(err)
					}
					return
				}
			}
		}()

		return &FFmpegDecoder{
			cmd:         cmd,
			stdout:      stdout,
			audioReader: audioReader,
			audioWriter: audioWriter,
			frameSize:   width * height * 3,
			width:       width,
			height:      height,
		}, nil

	case <-time.After(10 * time.Second):
		_ = cmd.Process.Kill()
		ln.Close()
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = "timeout sin logs de error"
		}
		return nil, fmt.Errorf("ffmpeg hang detected: %s", msg)
	}
}

func (d *FFmpegDecoder) AudioReader() io.Reader {
	if d == nil {
		return nil
	}
	return d.audioReader
}

func (d *FFmpegDecoder) Next(ctx context.Context) (types.FrameRGB, error) {
	if d == nil || d.stdout == nil {
		return types.FrameRGB{}, io.EOF
	}

	if ctx != nil {
		select {
		case <-ctx.Done():
			return types.FrameRGB{}, ctx.Err()
		default:
		}
	}

	if cap(d.buf) < d.frameSize {
		d.buf = make([]byte, d.frameSize)
	}
	buf := d.buf[:d.frameSize]

	if _, err := io.ReadFull(d.stdout, buf); err != nil {
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			return types.FrameRGB{}, io.EOF
		}
		if ctx != nil && ctx.Err() != nil {
			return types.FrameRGB{}, ctx.Err()
		}
		return types.FrameRGB{}, fmt.Errorf("read ffmpeg frame: %w", err)
	}

	return types.FrameRGB{
		W:      d.width,
		H:      d.height,
		Stride: d.width * 3,
		Pix:    buf,
	}, nil
}

func (d *FFmpegDecoder) Close() error {
	if d == nil {
		return nil
	}

	if d.stdout != nil {
		_ = d.stdout.Close()
	}

	if d.audioWriter != nil {
		_ = d.audioWriter.Close()
	}

	if d.audioReader != nil {
		_ = d.audioReader.Close()
	}

	if d.cmd != nil {
		_ = d.cmd.Wait()
	}

	return nil
}

func ProbeVideoSize(ctx context.Context, inputPath, ffprobePath string) (int, int, error) {
	cmd := exec.CommandContext(
		ctx,
		ffprobePath,
		"-v",
		"error",
		"-select_streams",
		"v:0",
		"-show_entries",
		"stream=width,height",
		"-of",
		"json",
		inputPath,
	)

	out, err := cmd.CombinedOutput()
	if err != nil {
		msg := strings.TrimSpace(string(out))
		if msg != "" {
			return 0, 0, fmt.Errorf("run ffprobe: %w: %s", err, msg)
		}
		return 0, 0, fmt.Errorf("run ffprobe: %w", err)
	}

	var payload ffprobeResponse
	if err := json.Unmarshal(out, &payload); err != nil {
		return 0, 0, fmt.Errorf("parse ffprobe output: %w", err)
	}

	if len(payload.Streams) == 0 || payload.Streams[0].Width <= 0 || payload.Streams[0].Height <= 0 {
		return 0, 0, fmt.Errorf("ffprobe returned invalid dimensions")
	}

	return payload.Streams[0].Width, payload.Streams[0].Height, nil
}