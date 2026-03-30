package term

import (
	"context"
	"os"
	"sync"
	"time"

	xterm "golang.org/x/term"
)

// WatchSeek escucha flechas izquierda/derecha y emite saltos temporales.
// Devuelve nil si stdin no es un terminal interactivo.
func WatchSeek(ctx context.Context, step time.Duration) <-chan time.Duration {
	if step <= 0 {
		return nil
	}

	fd := int(os.Stdin.Fd())
	if !xterm.IsTerminal(fd) {
		return nil
	}

	oldState, err := xterm.MakeRaw(fd)
	if err != nil {
		return nil
	}

	ch := make(chan time.Duration, 1)
	var once sync.Once
	restore := func() {
		once.Do(func() {
			_ = xterm.Restore(fd, oldState)
		})
	}

	go func() {
		defer restore()
		defer close(ch)

		buf := make([]byte, 1)
		state := 0

		send := func(delta time.Duration) {
			select {
			case ch <- delta:
			default:
				select {
				case <-ch:
				default:
				}
				select {
				case ch <- delta:
				default:
				}
			}
		}

		for {
			select {
			case <-ctx.Done():
				return
			default:
			}

			n, err := os.Stdin.Read(buf)
			if err != nil || n == 0 {
				return
			}

			b := buf[0]
			switch state {
			case 0:
				if b == 0x1b {
					state = 1
				}
			case 1:
				if b == '[' {
					state = 2
				} else {
					state = 0
				}
			case 2:
				switch b {
				case 'C':
					send(step)
				case 'D':
					send(-step)
				}
				state = 0
			}
		}
	}()

	return ch
}
