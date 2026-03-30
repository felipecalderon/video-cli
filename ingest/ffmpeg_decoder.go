package ingest

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os/exec"
	"time"
	"video-terminal/types"
)

type FFmpegDecoder struct {
	cmd       *exec.Cmd
	stdout    io.ReadCloser
	audioConn net.Conn
	frameSize int
	width     int
	height    int
	buf       []byte
}

type ffprobeResponse struct {
	Streams []struct {
		Width  int `json:"width"`
		Height int `json:"height"`
	} `json:"streams"`
}

func NewFFmpegDecoder(ctx context.Context, inputPath string, fps int, ffmpegPath, ffprobePath string, isStream bool, startOffset time.Duration) (*FFmpegDecoder, error) {
	if fps <= 0 {
		fps = 15
	}

	width, height, err := probeVideoSize(ctx, inputPath, ffprobePath)
	if err != nil {
		return nil, err
	}

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, fmt.Errorf("create audio listener: %w", err)
	}
	defer ln.Close()

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

	// Buffer para capturar el error real de FFmpeg si falla al arrancar
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("create ffmpeg stdout pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("start ffmpeg: %w", err)
	}

	// Canal para recibir la conexión o el error con timeout
	type acceptRes struct {
		conn net.Conn
		err  error
	}
	resChan := make(chan acceptRes, 1)
	go func() {
		conn, err := ln.Accept()
		resChan <- acceptRes{conn, err}
	}()

	select {
	case res := <-resChan:
		if res.err != nil {
			cmd.Process.Kill()
			return nil, fmt.Errorf("audio connection failed: %w", res.err)
		}
		return &FFmpegDecoder{
			cmd:       cmd,
			stdout:    stdout,
			audioConn: res.conn,
			frameSize: width * height * 3,
			width:     width,
			height:    height,
		}, nil
	case <-time.After(10 * time.Second): // 10 segundos de margen para conectar
		cmd.Process.Kill()
		msg := stderr.String()
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
	return d.audioConn
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

	if d.audioConn != nil {
		_ = d.audioConn.Close()
	}

	if d.cmd != nil {
		return d.cmd.Wait()
	}

	return nil
}

func probeVideoSize(ctx context.Context, inputPath, ffprobePath string) (int, int, error) {
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

	out, err := cmd.Output()
	if err != nil {
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
