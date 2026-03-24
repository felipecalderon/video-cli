package render

const (
	quantizeTruecolorLevels = 64
	quantizeAnsi256Levels   = 6
)

var bayer4x4 = [4][4]uint8{
	{0, 8, 2, 10},
	{12, 4, 14, 6},
	{3, 11, 1, 9},
	{15, 7, 13, 5},
}

var bayer8x8 = [8][8]uint8{
	{0, 48, 12, 60, 3, 51, 15, 63},
	{32, 16, 44, 28, 35, 19, 47, 31},
	{8, 56, 4, 52, 11, 59, 7, 55},
	{40, 24, 36, 20, 43, 27, 39, 23},
	{2, 50, 14, 62, 1, 49, 13, 61},
	{34, 18, 46, 30, 33, 17, 45, 29},
	{10, 58, 6, 54, 9, 57, 5, 53},
	{42, 26, 38, 22, 41, 25, 37, 21},
}

func quantizeChannel(c uint8, levels int) uint8 {
	if levels < 2 {
		return c
	}

	idx := int(c) * (levels - 1) / 255
	if idx < 0 {
		idx = 0
	}
	if idx > levels-1 {
		idx = levels - 1
	}

	if levels == 1 {
		return 0
	}

	return uint8((idx*255 + (levels-1)/2) / (levels - 1))
}

func ditherChannel(c uint8, threshold uint8, area int) uint8 {
	if area <= 0 {
		return c
	}

	biasRange := 2
	bias := int(threshold)*((biasRange*2)+1)/area - biasRange
	if bias == 0 {
		return c
	}

	value := int(c) + bias
	if value < 0 {
		value = 0
	} else if value > 255 {
		value = 255
	}

	return uint8(value)
}
