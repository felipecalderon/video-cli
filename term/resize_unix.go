//go:build !windows

package term

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

func WatchSize(ctx context.Context) <-chan struct{} {
	ch := make(chan struct{}, 1)
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGWINCH)

	go func() {
		defer signal.Stop(sigCh)
		for {
			select {
			case <-ctx.Done():
				return
			case <-sigCh:
				select {
				case <-ch:
				default:
				}
				ch <- struct{}{}
			}
		}
	}()
	return ch
}
