package ingest

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"video-terminal/types"
)

type FFmpegDecoder struct {
	cmd       *exec.Cmd
	stdout    io.ReadCloser
	frameSize int
	width     int
	height    int
}

type ffprobeResponse struct {
	Streams []struct {
		Width  int `json:"width"`
		Height int `json:"height"`
	} `json:"streams"`
}

func NewFFmpegDecoder(ctx context.Context, inputPath string, fps int, ffmpegPath, ffprobePath string) (*FFmpegDecoder, error) {
	if fps <= 0 {
		fps = 15
	}
	if strings.TrimSpace(ffmpegPath) == "" {
		return nil, fmt.Errorf("ffmpeg binary path is required")
	}
	if strings.TrimSpace(ffprobePath) == "" {
		return nil, fmt.Errorf("ffprobe binary path is required")
	}

	width, height, err := probeVideoSize(ctx, inputPath, ffprobePath)
	if err != nil {
		return nil, err
	}

	cmd := exec.CommandContext(
		ctx,
		ffmpegPath,
		"-hide_banner",
		"-loglevel",
		"error",
		"-i",
		inputPath,
		"-vf",
		fmt.Sprintf("fps=%d", fps),
		"-f",
		"rawvideo",
		"-pix_fmt",
		"rgb24",
		"pipe:1",
	)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("create ffmpeg stdout pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("start ffmpeg: %w", err)
	}

	return &FFmpegDecoder{
		cmd:       cmd,
		stdout:    stdout,
		frameSize: width * height * 3,
		width:     width,
		height:    height,
	}, nil
}

func (d *FFmpegDecoder) Next(ctx context.Context) (types.FrameRGB, error) {
	if d == nil || d.stdout == nil {
		return types.FrameRGB{}, io.EOF
	}

	buf := make([]byte, d.frameSize)

	errCh := make(chan error, 1)
	go func() {
		_, err := io.ReadFull(d.stdout, buf)
		errCh <- err
	}()

	select {
	case <-ctx.Done():
		return types.FrameRGB{}, ctx.Err()
	case err := <-errCh:
		if err != nil {
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				return types.FrameRGB{}, io.EOF
			}
			return types.FrameRGB{}, fmt.Errorf("read ffmpeg frame: %w", err)
		}
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
