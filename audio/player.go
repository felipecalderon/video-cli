package audio

import (
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/ebitengine/oto/v3"
)

// Player se encarga de la reproducción de audio y de actuar como Master Clock.
type Player struct {
	context *oto.Context
	player  *oto.Player

	mu      sync.RWMutex
	started time.Time
}

// NewPlayer inicializa el motor de audio OTO.
func NewPlayer() (*Player, error) {
	op := &oto.NewContextOptions{
		SampleRate:   44100,
		ChannelCount: 2,
		Format:       oto.FormatSignedInt16LE,
	}

	ctx, ready, err := oto.NewContext(op)
	if err != nil {
		return nil, fmt.Errorf("oto new context: %w", err)
	}
	<-ready // Esperar a que el dispositivo esté listo

	return &Player{
		context: ctx,
	}, nil
}

// Play inicia la reproducción desde un io.Reader PCM s16le.
func (p *Player) Play(r io.Reader) error {
	p.player = p.context.NewPlayer(r)
	p.player.Play()

	p.mu.Lock()
	p.started = time.Now()
	p.mu.Unlock()

	return nil
}
func (p *Player) CurrentTime() time.Duration {
	p.mu.RLock()
	defer p.mu.RUnlock()
	if p.started.IsZero() {
		return 0
	}
	return time.Since(p.started)
}

func (p *Player) Close() error {
	if p.player != nil {
		return p.player.Close()
	}
	return nil
}
