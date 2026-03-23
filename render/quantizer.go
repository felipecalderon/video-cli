package render

import (
	"context"
	"video-terminal/types"
)

type NoopQuantizer struct{}

func (NoopQuantizer) Quantize(ctx context.Context, in types.WorkRGB, mode types.ColorMode) (types.WorkRGB, error) {
	_ = ctx
	_ = mode
	return in, nil
}
