package render

import (
	"bufio"
	"context"
	"io"
	"strconv"
	"unicode/utf8"
	"video-terminal/types"
)

type ANSIOutput struct {
	writer     *bufio.Writer
	colorMode  types.ColorMode
	buf        []byte
	lastFG     [3]uint8
	lastBG     [3]uint8
	lastANSIFG int
	lastANSIBG int
	hasColor   bool
}

func NewANSIOutput(w io.Writer, mode types.ColorMode) *ANSIOutput {
	return &ANSIOutput{
		writer:     bufio.NewWriterSize(w, 64*1024),
		colorMode:  mode,
		lastANSIFG: -1,
		lastANSIBG: -1,
	}
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

	curX := -1
	curY := -1

	for _, op := range ops {
		// 1. Cursor movement optimization: omit or shorten when on same row
		if curY == op.Y && curX == op.X {
			// Cursor is already at the target position
		} else if curY == op.Y && op.X > curX {
			gap := op.X - curX
			if gap == 1 {
				buf = append(buf, '\x1b', '[', 'C')
			} else {
				buf = append(buf, '\x1b', '[')
				buf = strconv.AppendInt(buf, int64(gap), 10)
				buf = append(buf, 'C')
			}
		} else {
			buf = append(buf, '\x1b', '[')
			buf = strconv.AppendInt(buf, int64(op.Y+1), 10)
			buf = append(buf, ';')
			buf = strconv.AppendInt(buf, int64(op.X+1), 10)
			buf = append(buf, 'H')
		}

		// 2. Color emission optimization: only emit components that actually changed
		if o.colorMode == types.Color256 {
			fg := rgbToANSI256(op.FG)
			bg := rgbToANSI256(op.BG)
			fgChanged := !o.hasColor || o.lastANSIFG != fg
			bgChanged := !o.hasColor || o.lastANSIBG != bg

			if fgChanged && bgChanged {
				buf = append(buf, '\x1b', '[')
				buf = append(buf, "38;5;"...)
				buf = strconv.AppendInt(buf, int64(fg), 10)
				buf = append(buf, ';', '4', '8', ';', '5', ';')
				buf = strconv.AppendInt(buf, int64(bg), 10)
				buf = append(buf, 'm')
			} else if fgChanged {
				buf = append(buf, '\x1b', '[')
				buf = append(buf, "38;5;"...)
				buf = strconv.AppendInt(buf, int64(fg), 10)
				buf = append(buf, 'm')
			} else if bgChanged {
				buf = append(buf, '\x1b', '[')
				buf = append(buf, "48;5;"...)
				buf = strconv.AppendInt(buf, int64(bg), 10)
				buf = append(buf, 'm')
			}
			o.lastANSIFG = fg
			o.lastANSIBG = bg
			o.lastFG = op.FG
			o.lastBG = op.BG
			o.hasColor = true
		} else {
			fgChanged := !o.hasColor || o.lastFG != op.FG
			bgChanged := !o.hasColor || o.lastBG != op.BG

			if fgChanged && bgChanged {
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
			} else if fgChanged {
				buf = append(buf, '\x1b', '[')
				buf = append(buf, "38;2;"...)
				buf = strconv.AppendInt(buf, int64(op.FG[0]), 10)
				buf = append(buf, ';')
				buf = strconv.AppendInt(buf, int64(op.FG[1]), 10)
				buf = append(buf, ';')
				buf = strconv.AppendInt(buf, int64(op.FG[2]), 10)
				buf = append(buf, 'm')
			} else if bgChanged {
				buf = append(buf, '\x1b', '[')
				buf = append(buf, "48;2;"...)
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

		// 3. Characters output and cursor tracker update
		cols := 1
		if len(op.Text) > 0 {
			for _, r := range op.Text {
				buf = appendRuneUTF8(buf, r)
			}
			cols = len(op.Text)
		} else if op.RunLen > 0 {
			var rBuf [4]byte
			n := utf8.EncodeRune(rBuf[:], op.Ch)
			chBytes := rBuf[:n]
			for i := 0; i < op.RunLen; i++ {
				buf = append(buf, chBytes...)
			}
			cols = op.RunLen
		} else {
			buf = appendRuneUTF8(buf, op.Ch)
			cols = 1
		}

		curY = op.Y
		curX = op.X + cols
	}

	if _, err := o.writer.Write(buf); err != nil {
		return err
	}
	o.buf = buf

	return o.writer.Flush()
}

func rgbToANSI256(c [3]uint8) int {
	r, g, b := int(c[0]), int(c[1]), int(c[2])

	maxDiff := absInt(r - g)
	if d := absInt(g - b); d > maxDiff {
		maxDiff = d
	}
	if d := absInt(r - b); d > maxDiff {
		maxDiff = d
	}

	if maxDiff <= 8 {
		gray := (77*r + 150*g + 29*b) >> 8
		if gray < 4 {
			return 16
		}
		if gray > 246 {
			return 231
		}
		idx := (gray - 3) / 10
		if idx < 0 {
			idx = 0
		} else if idx > 23 {
			idx = 23
		}
		return 232 + idx
	}

	rIdx := (r*5 + 127) / 255
	gIdx := (g*5 + 127) / 255
	bIdx := (b*5 + 127) / 255
	return 16 + 36*rIdx + 6*gIdx + bIdx
}

func appendRuneUTF8(dst []byte, r rune) []byte {
	return utf8.AppendRune(dst, r)
}

func (o *ANSIOutput) Clear(ctx context.Context) error {
	_ = ctx
	if o == nil || o.writer == nil {
		return nil
	}
	o.hasColor = false
	o.lastFG = [3]uint8{}
	o.lastBG = [3]uint8{}
	o.lastANSIFG = -1
	o.lastANSIBG = -1
	if _, err := o.writer.WriteString("\x1b[0m\x1b[2J\x1b[H"); err != nil {
		return err
	}
	return o.writer.Flush()
}
