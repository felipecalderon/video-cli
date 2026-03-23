package render

import (
	"context"
	"fmt"
	"io"
	"video-terminal/types"
)

type ANSIOutput struct {
	writer    io.Writer
	colorMode types.ColorMode
}

func NewANSIOutput(w io.Writer, mode types.ColorMode) ANSIOutput {
	return ANSIOutput{writer: w, colorMode: mode}
}

func (o ANSIOutput) Write(ctx context.Context, ops []types.DiffOp) error {
	_ = ctx

	if len(ops) == 0 {
		return nil
	}

	buf := make([]byte, 0, len(ops)*40)
	for _, op := range ops {
		move := fmt.Sprintf("\x1b[%d;%dH", op.Y+1, op.X+1)
		buf = append(buf, move...)

		if o.colorMode == types.Color256 {
			fg := rgbToANSI256(op.FG)
			bg := rgbToANSI256(op.BG)
			color := fmt.Sprintf("\x1b[38;5;%d;48;5;%dm", fg, bg)
			buf = append(buf, color...)
		} else {
			color := fmt.Sprintf(
				"\x1b[38;2;%d;%d;%d;48;2;%d;%d;%dm",
				op.FG[0], op.FG[1], op.FG[2],
				op.BG[0], op.BG[1], op.BG[2],
			)
			buf = append(buf, color...)
		}

		buf = appendRuneUTF8(buf, op.Ch)
	}

	_, err := o.writer.Write(buf)
	return err
}

func rgbToANSI256(c [3]uint8) int {
	r := int(c[0]) * 5 / 255
	g := int(c[1]) * 5 / 255
	b := int(c[2]) * 5 / 255
	return 16 + 36*r + 6*g + b
}

func appendRuneUTF8(dst []byte, r rune) []byte {
	if r < 0x80 {
		return append(dst, byte(r))
	}
	return append(dst, string(r)...)
}
