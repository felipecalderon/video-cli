package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"math"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"video-terminal/audio"
	"video-terminal/diff"
	"video-terminal/ingest"
	"video-terminal/pipeline"
	"video-terminal/render"
	"video-terminal/term"
	"video-terminal/types"
)

type stringFlag struct {
	set bool
	v   *string
}

func (f *stringFlag) String() string {
	if f == nil || f.v == nil {
		return ""
	}
	return *f.v
}

func (f *stringFlag) Set(s string) error {
	if f == nil || f.v == nil {
		return fmt.Errorf("string flag not initialized")
	}
	f.set = true
	*f.v = s
	return nil
}

type intFlag struct {
	set bool
	v   *int
}

func (f *intFlag) String() string {
	if f == nil || f.v == nil {
		return ""
	}
	return strconv.Itoa(*f.v)
}

func (f *intFlag) Set(s string) error {
	if f == nil || f.v == nil {
		return fmt.Errorf("int flag not initialized")
	}
	val, err := strconv.Atoi(s)
	if err != nil {
		return fmt.Errorf("invalid int value %q", s)
	}
	f.set = true
	*f.v = val
	return nil
}

type floatFlag struct {
	set bool
	v   *float64
}

func (f *floatFlag) String() string {
	if f == nil || f.v == nil {
		return ""
	}
	return strconv.FormatFloat(*f.v, 'f', -1, 64)
}

func (f *floatFlag) Set(s string) error {
	if f == nil || f.v == nil {
		return fmt.Errorf("float flag not initialized")
	}
	val, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return fmt.Errorf("invalid float value %q", s)
	}
	f.set = true
	*f.v = val
	return nil
}

type durationFlag struct {
	set bool
	v   *time.Duration
}

func (f *durationFlag) String() string {
	if f == nil || f.v == nil {
		return ""
	}
	return f.v.String()
}

func (f *durationFlag) Set(s string) error {
	if f == nil || f.v == nil {
		return fmt.Errorf("duration flag not initialized")
	}
	d, err := time.ParseDuration(s)
	if err != nil {
		return fmt.Errorf("invalid duration value %q", s)
	}
	if d <= 0 {
		return fmt.Errorf("duration must be > 0")
	}
	f.set = true
	*f.v = d
	return nil
}

type playerConfig struct {
	Fps        *int     `json:"fps"`
	Preset     *string  `json:"preset"`
	Color      *string  `json:"color"`
	Scale      *float64 `json:"scale"`
	TermWidth  *int     `json:"term_width"`
	TermHeight *int     `json:"term_height"`
	BlendAlpha *float64 `json:"blend_alpha"`
	SeekStep   *string  `json:"seek_step"`
}

func loadConfig(path string) (playerConfig, error) {
	if strings.TrimSpace(path) == "" {
		return playerConfig{}, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return playerConfig{}, fmt.Errorf("read config: %w", err)
	}

	var cfg playerConfig
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&cfg); err != nil {
		return playerConfig{}, fmt.Errorf("parse config json: %w", err)
	}
	if err := dec.Decode(&struct{}{}); err != io.EOF {
		if err == nil {
			return playerConfig{}, fmt.Errorf("parse config json: extra data after JSON object")
		}
		return playerConfig{}, fmt.Errorf("parse config json: %w", err)
	}
	if err := validateConfig(&cfg); err != nil {
		return playerConfig{}, err
	}

	return cfg, nil
}

func validateConfig(cfg *playerConfig) error {
	if cfg == nil {
		return nil
	}

	if cfg.Fps != nil && *cfg.Fps <= 0 {
		return fmt.Errorf("config fps must be > 0")
	}
	if cfg.Scale != nil && *cfg.Scale <= 0 {
		return fmt.Errorf("config scale must be > 0")
	}
	if cfg.TermWidth != nil && *cfg.TermWidth <= 0 {
		return fmt.Errorf("config term_width must be > 0")
	}
	if cfg.TermHeight != nil && *cfg.TermHeight <= 0 {
		return fmt.Errorf("config term_height must be > 0")
	}
	if cfg.BlendAlpha != nil {
		if *cfg.BlendAlpha < 0 || *cfg.BlendAlpha > 1 {
			return fmt.Errorf("config blend_alpha must be between 0 and 1")
		}
	}
	if cfg.SeekStep != nil {
		d, err := time.ParseDuration(strings.TrimSpace(*cfg.SeekStep))
		if err != nil {
			return fmt.Errorf("config seek_step must be a valid duration")
		}
		if d <= 0 {
			return fmt.Errorf("config seek_step must be > 0")
		}
		*cfg.SeekStep = d.String()
	}
	if cfg.Preset != nil {
		v := strings.ToLower(strings.TrimSpace(*cfg.Preset))
		if !isValidPreset(v) {
			return fmt.Errorf("config preset must be one of: fast, quality, crt")
		}
		*cfg.Preset = v
	}
	if cfg.Color != nil {
		v := strings.ToLower(strings.TrimSpace(*cfg.Color))
		if !isValidColor(v) {
			return fmt.Errorf("config color must be one of: auto, truecolor, 256")
		}
		*cfg.Color = v
	}

	return nil
}

func main() {
	input := flag.String("input", "", "Path to video file")
	configPath := flag.String("config", "", "Path to JSON config file (default: ./config.json if present)")

	fps := 15
	color := "auto"
	preset := "fast"
	scale := 1.0
	termWOverride := 0
	termHOverride := 0
	blendAlpha := 0.0
	blendAlphaSet := false
	seekStep := 5 * time.Second

	fpsFlag := intFlag{v: &fps}
	colorFlag := stringFlag{v: &color}
	presetFlag := stringFlag{v: &preset}
	scaleFlag := floatFlag{v: &scale}
	termWFlag := intFlag{v: &termWOverride}
	termHFlag := intFlag{v: &termHOverride}
	blendAlphaFlag := floatFlag{v: &blendAlpha}
	seekStepFlag := durationFlag{v: &seekStep}

	flag.Var(&fpsFlag, "fps", "Target FPS")
	flag.Var(&colorFlag, "color", "Color mode: auto|truecolor|256")
	flag.Var(&presetFlag, "preset", "Preset: fast|quality|crt")
	flag.Var(&scaleFlag, "scale", "Scale multiplier for detected terminal size (e.g. 0.8)")
	flag.Var(&termWFlag, "term-width", "Override terminal width (columns)")
	flag.Var(&termHFlag, "term-height", "Override terminal height (rows)")
	flag.Var(&blendAlphaFlag, "blend-alpha", "Temporal blend alpha (0..1)")
	flag.Var(&seekStepFlag, "seek-step", "Seek step for left/right arrows (e.g. 5s)")

	ffmpegPath := flag.String("ffmpeg", "", "Path to ffmpeg binary")
	ffprobePath := flag.String("ffprobe", "", "Path to ffprobe binary")
	flag.Parse()

	if strings.TrimSpace(*input) == "" {
		slog.Error("missing required --input")
		os.Exit(2)
	}

	cfgPath := strings.TrimSpace(*configPath)
	if cfgPath == "" {
		if _, err := os.Stat("config.json"); err == nil {
			cfgPath = "config.json"
		}
	}

	cfg, err := loadConfig(cfgPath)
	if err != nil {
		slog.Error("failed to load config", "err", err)
		os.Exit(2)
	}

	if cfg.Fps != nil {
		fps = *cfg.Fps
	}
	if cfg.Color != nil {
		color = *cfg.Color
	}
	if cfg.Preset != nil {
		preset = *cfg.Preset
	}
	if cfg.Scale != nil {
		scale = *cfg.Scale
	}
	if cfg.TermWidth != nil {
		termWOverride = *cfg.TermWidth
	}
	if cfg.TermHeight != nil {
		termHOverride = *cfg.TermHeight
	}
	if cfg.BlendAlpha != nil {
		blendAlpha = *cfg.BlendAlpha
		blendAlphaSet = true
	}
	if cfg.SeekStep != nil {
		d, err := time.ParseDuration(*cfg.SeekStep)
		if err != nil {
			slog.Error("invalid seek_step in config", "err", err)
			os.Exit(2)
		}
		seekStep = d
	}

	if fpsFlag.set {
		fps = *fpsFlag.v
	}
	if colorFlag.set {
		color = *colorFlag.v
	}
	if presetFlag.set {
		preset = *presetFlag.v
	}
	if scaleFlag.set {
		scale = *scaleFlag.v
	}
	if termWFlag.set {
		termWOverride = *termWFlag.v
	}
	if termHFlag.set {
		termHOverride = *termHFlag.v
	}
	if blendAlphaFlag.set {
		blendAlpha = *blendAlphaFlag.v
		blendAlphaSet = true
	}
	if seekStepFlag.set {
		seekStep = *seekStepFlag.v
	}

	if fps <= 0 {
		slog.Error("invalid fps (must be > 0)")
		os.Exit(2)
	}
	if scale <= 0 {
		slog.Error("invalid scale (must be > 0)")
		os.Exit(2)
	}
	if !isValidPreset(preset) {
		slog.Error("invalid preset", "value", preset)
		os.Exit(2)
	}
	if !isValidColor(color) {
		slog.Error("invalid color", "value", color)
		os.Exit(2)
	}
	if !blendAlphaSet {
		blendAlpha = defaultBlendForPreset(preset)
	}
	if blendAlpha < 0 || blendAlpha > 1 {
		slog.Error("invalid blend-alpha (must be 0..1)")
		os.Exit(2)
	}
	if seekStep <= 0 {
		slog.Error("invalid seek-step (must be > 0)")
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

	// --- Resolución automática de URLs remotas ---
	isStream := ingest.IsURL(*input)
	if isStream {
		resolver := &ingest.YtdlpResolver{} // busca yt-dlp en PATH
		result, err := resolver.Resolve(ctx, *input, 480)
		if err != nil {
			slog.Error("failed to resolve stream URL", "resolver", resolver.Name(), "err", err)
			os.Exit(1)
		}
		if result.Title != "" {
			slog.Info("stream resolved", "title", result.Title)
		}
		*input = result.URL
	}

	termW, termH := computeTermSize(termWOverride, termHOverride, scale)

	resizeEvents := make(chan [2]int, 1)
	resizeSig := term.WatchSize(ctx)
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-resizeSig:
				newW, newH := computeTermSize(termWOverride, termHOverride, scale)
				select {
				case <-resizeEvents:
				default:
				}
				resizeEvents <- [2]int{newW, newH}
			}
		}
	}()
	seekCh := term.WatchSeek(ctx, seekStep)

	mode := term.ResolveColorMode(color)
	output := render.NewANSIOutput(os.Stdout, mode)

	fmt.Print("\x1b[2J\x1b[H\x1b[?25l")
	defer fmt.Print("\x1b[0m\x1b[?25h\n")

	baseOffset := time.Duration(0)
	for {
		seekDelta, shouldSeek, err := runPlaybackSession(ctx, playbackSessionParams{
			Input:       *input,
			BaseOffset:  baseOffset,
			FPS:         fps,
			Mode:        mode,
			Preset:      parsePreset(preset),
			BlendAlpha:  blendAlpha,
			TermW:       termW,
			TermH:       termH,
			FFmpegPath:  resolvedFFmpeg,
			FFprobePath: resolvedFFprobe,
			IsStream:    isStream,
			ResizeChan:  resizeEvents,
			SeekCh:      seekCh,
			SeekStep:    seekStep,
			Output:      output,
		})
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return
			}
			slog.Error("playback session failed", "err", err)
			os.Exit(1)
		}
		if shouldSeek {
			baseOffset += seekDelta
			if baseOffset < 0 {
				baseOffset = 0
			}
			fmt.Print("\x1b[2J\x1b[H")
			continue
		}
		return
	}
}

type playbackSessionParams struct {
	Input       string
	BaseOffset  time.Duration
	FPS         int
	Mode        types.ColorMode
	Preset      types.Preset
	BlendAlpha  float64
	TermW       int
	TermH       int
	FFmpegPath  string
	FFprobePath string
	IsStream    bool
	ResizeChan  <-chan [2]int
	SeekCh      <-chan time.Duration
	SeekStep    time.Duration
	Output      pipeline.Output
}

func runPlaybackSession(ctx context.Context, params playbackSessionParams) (time.Duration, bool, error) {
	sessionCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	decoder, err := ingest.NewFFmpegDecoder(sessionCtx, params.Input, params.FPS, params.FFmpegPath, params.FFprobePath, params.IsStream, params.BaseOffset)
	if err != nil {
		return 0, false, err
	}
	defer func() {
		if err := decoder.Close(); err != nil && !errors.Is(err, io.EOF) {
			slog.Warn("decoder close error", "err", err)
		}
	}()

	var audioClock types.Clock
	audioPlayer, err := audio.NewPlayer()
	if err != nil {
		slog.Warn("audio player initialization failed, continuing without sound", "err", err)
	} else {
		defer audioPlayer.Close()
		if err := audioPlayer.Play(decoder.AudioReader()); err != nil {
			slog.Warn("audio playback start failed", "err", err)
		} else {
			audioClock = audioPlayer
		}
	}

	p := pipeline.Pipeline{
		Decoder:   decoder,
		Resizer:   &render.NearestResizer{},
		Quantizer: render.ChannelQuantizer{},
		Dither:    &render.BayerDither{},
		Temporal:  &render.TemporalBlend{},
		Scanliner: render.ScanlineEffect{},
		Mapper:    render.BlockMapper{},
		Differ:    &diff.ByteDiffer{},
		Output:    params.Output,
	}

	runParams := types.PipelineParams{
		TermW:      params.TermW,
		TermH:      params.TermH,
		FpsTarget:  params.FPS,
		ColorMode:  params.Mode,
		Preset:     params.Preset,
		BlendAlpha: params.BlendAlpha,
		ResizeChan: params.ResizeChan,
		Clock:      audioClock,
	}

	done := make(chan error, 1)
	go func() {
		done <- p.Run(sessionCtx, runParams)
	}()

	if params.SeekCh == nil {
		err = <-done
		return 0, false, err
	}

	for {
		select {
		case <-ctx.Done():
			cancel()
			<-done
			return 0, false, ctx.Err()
		case delta, ok := <-params.SeekCh:
			if !ok {
				params.SeekCh = nil
				continue
			}
			cancel()
			<-done
			return delta, true, nil
		case err = <-done:
			return 0, false, err
		}
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

func defaultBlendForPreset(v string) float64 {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "quality":
		return 0.18
	case "crt":
		return 0.38
	default:
		return 0
	}
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

func isValidPreset(v string) bool {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "fast", "quality", "crt":
		return true
	default:
		return false
	}
}

func isValidColor(v string) bool {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "auto", "truecolor", "256":
		return true
	default:
		return false
	}
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func computeTermSize(termWOverride, termHOverride int, scale float64) (int, int) {
	actualW, actualH := term.GetSize()
	termW := actualW
	termH := actualH

	if termWOverride > 0 {
		termW = minInt(termWOverride, actualW)
	}
	if termHOverride > 0 {
		termH = minInt(termHOverride, actualH)
	}

	if scale != 1 {
		termW = int(math.Round(float64(termW) * scale))
		termH = int(math.Round(float64(termH) * scale))
	}

	if termW < 1 {
		termW = 1
	}
	if termH < 1 {
		termH = 1
	}
	if termW > actualW {
		termW = actualW
	}
	if termH > actualH {
		termH = actualH
	}
	return termW, termH
}
