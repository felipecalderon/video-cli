package types

type FrameRGB struct {
	W, H   int
	Stride int
	Pix    []uint8
}

type WorkRGB struct {
	W, H   int
	Stride int
	Pix    []uint8
}

type Cell struct {
	Top    [3]uint8
	Bottom [3]uint8
	Ch     rune
}

type CellGrid struct {
	W, H  int
	Cells []Cell
}

type DiffOp struct {
	X, Y int
	FG   [3]uint8
	BG   [3]uint8
	Ch   rune
	Text string
}

type ColorMode uint8

const (
	ColorAuto ColorMode = iota
	ColorTruecolor
	Color256
)

type Preset uint8

const (
	PresetFast Preset = iota
	PresetQuality
	PresetCRT
)

type PipelineParams struct {
	TermW, TermH int
	FpsTarget    int
	ColorMode    ColorMode
	Preset       Preset
}
