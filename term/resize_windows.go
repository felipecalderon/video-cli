//go:build windows

package term

import (
	"context"
	"time"
)

func WatchSize(ctx context.Context) <-chan struct{} {
	ch := make(chan struct{}, 1)
	go func() {
		lastW, lastH := GetSize()
		ticker := time.NewTicker(500 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				w, h := GetSize()
				if w != lastW || h != lastH {
					lastW, lastH = w, h
					// Ensure latest event is queued
					select {
					case <-ch:
					default:
					}
					ch <- struct{}{}
				}
			}
		}
	}()
	return ch
}
