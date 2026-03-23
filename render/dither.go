package render

import (
	"context"
	"video-terminal/types"
)

type NoopDither struct{}

func (NoopDither) Dither(ctx context.Context, in types.WorkRGB, preset types.Preset) (types.WorkRGB, error) {
	_ = ctx
	_ = preset
	return in, nil
}
