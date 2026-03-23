package term

import (
	"os"
	"strings"
	"video-terminal/types"

	xterm "golang.org/x/term"
)

func GetSize() (int, int) {
	w, h, err := xterm.GetSize(int(os.Stdout.Fd()))
	if err != nil || w <= 0 || h <= 0 {
		return 120, 40
	}
	return w, h
}

func ResolveColorMode(mode string) types.ColorMode {
	switch strings.ToLower(strings.TrimSpace(mode)) {
	case "truecolor":
		return types.ColorTruecolor
	case "256":
		return types.Color256
	default:
		return detectColorModeFromEnv()
	}
}

func detectColorModeFromEnv() types.ColorMode {
	ct := strings.ToLower(os.Getenv("COLORTERM"))
	if strings.Contains(ct, "truecolor") || strings.Contains(ct, "24bit") {
		return types.ColorTruecolor
	}

	t := strings.ToLower(os.Getenv("TERM"))
	if strings.Contains(t, "256color") {
		return types.Color256
	}

	return types.ColorTruecolor
}
