package audio

import (
	"encoding/binary"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/ebitengine/oto/v3"
	"github.com/gopxl/beep"
)

// PCMStreamer lee datos raw int16 de un io.Reader y los convierte en muestras para Beep.
type PCMStreamer struct {
	r io.Reader
	f beep.Format
}

func (s *PCMStreamer) Stream(samples [][2]float64) (n int, ok bool) {
	var l, r int16
	for i := range samples {
		// Leemos canal izquierdo (int16 Little Endian)
		if err := binary.Read(s.r, binary.LittleEndian, &l); err != nil {
			return i, i > 0
		}
		// Leemos canal derecho (int16 Little Endian)
		if err := binary.Read(s.r, binary.LittleEndian, &r); err != nil {
			return i, i > 0
		}
		// Convertimos a float64 [-1.0, 1.0]
		samples[i][0] = float64(l) / 32768.0
		samples[i][1] = float64(r) / 32768.0
		n++
	}
	return n, true
}

func (s *PCMStreamer) Err() error { return nil }

// Player se encarga de la reproducción de audio y de actuar como Master Clock.
type Player struct {
	context *oto.Context
	player  *oto.Player
	
	sampleRate beep.SampleRate
	format     beep.Format

	mu       sync.RWMutex
	currTime time.Duration
	started  time.Time
}

// NewPlayer inicializa el motor de audio OTO.
func NewPlayer() (*Player, error) {
	sr := beep.SampleRate(44100)
	format := beep.Format{
		SampleRate:  sr,
		NumChannels: 2,
		Precision:   2, // 16-bit (2 bytes)
	}

	// Inicializar OTO con 44.1kHz y estéreo.
	op := &oto.NewContextOptions{
		SampleRate:   int(sr),
		ChannelCount: 2,
		Format:       oto.FormatSignedInt16LE,
	}

	ctx, ready, err := oto.NewContext(op)
	if err != nil {
		return nil, fmt.Errorf("oto new context: %w", err)
	}
	<-ready // Esperar a que el dispositivo esté listo

	return &Player{
		context:    ctx,
		sampleRate: sr,
		format:     format,
	}, nil
}

// Play inicia la reproducción desde un io.Reader (que viene de FFmpeg).
// Se ejecuta en su propia goroutine para no bloquear.
func (p *Player) Play(r io.Reader) error {
	// Usamos nuestro Streamer manual para interpretar el stream PCM s16le.
	streamer := &PCMStreamer{r: r, f: p.format}

	p.player = p.context.NewPlayer(p.decodeToReader(streamer))
	p.player.Play()

	p.mu.Lock()
	p.started = time.Now()
	p.mu.Unlock()

	return nil
}


// decodeToReader es un adaptador simple para pasar de beep.Streamer a io.Reader para Oto.
func (p *Player) decodeToReader(s beep.Streamer) io.Reader {
	return &streamReader{s: s, p: p}
}

type streamReader struct {
	s       beep.Streamer
	p       *Player
	samples [][2]float64 // Buffer reutilizable para evitar allocs constantes
}

func (r *streamReader) Read(buf []byte) (int, error) {
	// Calculamos cuántas muestras necesitamos para llenar el buffer de bytes (4 bytes por muestra stéreo int16)
	needed := len(buf) / 4
	if cap(r.samples) < needed {
		r.samples = make([][2]float64, needed)
	}
	samples := r.samples[:needed]

	n, ok := r.s.Stream(samples)
	if !ok {
		if err := r.s.Err(); err != nil {
			return 0, err
		}
		return 0, io.EOF
	}

	// Convertir floats de beep de vuelta a s16le para oto (sin allocs adicionales)
	for i := 0; i < n; i++ {
		l := int16(samples[i][0] * 32767)
		r := int16(samples[i][1] * 32767)
		buf[i*4] = byte(l)
		buf[i*4+1] = byte(l >> 8)
		buf[i*4+2] = byte(r)
		buf[i*4+3] = byte(r >> 8)
	}

	// Actualizar tiempo del Master Clock
	r.p.mu.Lock()
	r.p.currTime += r.p.sampleRate.D(n)
	r.p.mu.Unlock()

	return n * 4, nil
}


func (p *Player) CurrentTime() time.Duration {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.currTime
}

func (p *Player) Close() error {
	if p.player != nil {
		return p.player.Close()
	}
	return nil
}
