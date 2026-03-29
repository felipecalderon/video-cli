package render

import (
	"bufio"
	"context"
	"io"
	"strconv"
	"video-terminal/types"
)

type ANSIOutput struct {
	writer    *bufio.Writer
	colorMode types.ColorMode
	buf       []byte
	lastFG    [3]uint8
	lastBG    [3]uint8
	hasColor  bool
}

func NewANSIOutput(w io.Writer, mode types.ColorMode) *ANSIOutput {
	return &ANSIOutput{writer: bufio.NewWriterSize(w, 64*1024), colorMode: mode}
}

func (o *ANSIOutput) Write(ctx context.Context, ops []types.DiffOp) error {
	_ = ctx

	if len(ops) == 0 {
		return nil
	}

	if o == nil || o.writer == nil {
		return nil
	}
	need := len(ops) * 48
	if cap(o.buf) < need {
		o.buf = make([]byte, 0, need)
	}
	buf := o.buf[:0]
	for _, op := range ops {
		buf = append(buf, '\x1b', '[')
		buf = strconv.AppendInt(buf, int64(op.Y+1), 10)
		buf = append(buf, ';')
		buf = strconv.AppendInt(buf, int64(op.X+1), 10)
		buf = append(buf, 'H')

		if !o.hasColor || o.lastFG != op.FG || o.lastBG != op.BG {
			if o.colorMode == types.Color256 {
				fg := rgbToANSI256(op.FG)
				bg := rgbToANSI256(op.BG)
				buf = append(buf, '\x1b', '[')
				buf = append(buf, "38;5;"...)
				buf = strconv.AppendInt(buf, int64(fg), 10)
				buf = append(buf, ';', '4', '8', ';', '5', ';')
				buf = strconv.AppendInt(buf, int64(bg), 10)
				buf = append(buf, 'm')
			} else {
				buf = append(buf, '\x1b', '[')
				buf = append(buf, "38;2;"...)
				buf = strconv.AppendInt(buf, int64(op.FG[0]), 10)
				buf = append(buf, ';')
				buf = strconv.AppendInt(buf, int64(op.FG[1]), 10)
				buf = append(buf, ';')
				buf = strconv.AppendInt(buf, int64(op.FG[2]), 10)
				buf = append(buf, ';', '4', '8', ';', '2', ';')
				buf = strconv.AppendInt(buf, int64(op.BG[0]), 10)
				buf = append(buf, ';')
				buf = strconv.AppendInt(buf, int64(op.BG[1]), 10)
				buf = append(buf, ';')
				buf = strconv.AppendInt(buf, int64(op.BG[2]), 10)
				buf = append(buf, 'm')
			}
			o.lastFG = op.FG
			o.lastBG = op.BG
			o.hasColor = true
		}

		if len(op.Text) > 0 {
			for _, r := range op.Text {
				buf = appendRuneUTF8(buf, r)
			}
		} else {
			buf = appendRuneUTF8(buf, op.Ch)
		}
	}

	if _, err := o.writer.Write(buf); err != nil {
		return err
	}
	o.buf = buf

	return o.writer.Flush()
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

func (o *ANSIOutput) Clear(ctx context.Context) error {
	_ = ctx
	if o == nil || o.writer == nil {
		return nil
	}
	o.hasColor = false
	o.lastFG = [3]uint8{}
	o.lastBG = [3]uint8{}
	if _, err := o.writer.WriteString("\x1b[0m\x1b[2J\x1b[H"); err != nil {
		return err
	}
	return o.writer.Flush()
}
