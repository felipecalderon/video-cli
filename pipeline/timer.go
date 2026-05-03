package pipeline

import (
	"sort"
	"sync/atomic"
	"time"
)

const ringSize = 120

type StageTimer struct {
	durations [ringSize]int64 // store as nanoseconds
	index     uint32
}

func (s *StageTimer) Mark(d time.Duration) {
	idx := atomic.AddUint32(&s.index, 1) % ringSize
	atomic.StoreInt64(&s.durations[idx], int64(d))
}

type StageMetrics struct {
	P50 time.Duration
	P95 time.Duration
	P99 time.Duration
	Avg time.Duration
}

func (s *StageTimer) Metrics() StageMetrics {
	var snaps [ringSize]int64
	var sum int64
	var count int
	for i := 0; i < ringSize; i++ {
		val := atomic.LoadInt64(&s.durations[i])
		if val > 0 {
			snaps[count] = val
			sum += val
			count++
		}
	}
	if count == 0 {
		return StageMetrics{}
	}

	validSnaps := snaps[:count]
	sort.Slice(validSnaps, func(i, j int) bool { return validSnaps[i] < validSnaps[j] })

	return StageMetrics{
		P50: time.Duration(validSnaps[count*50/100]),
		P95: time.Duration(validSnaps[count*95/100]),
		P99: time.Duration(validSnaps[count*99/100]),
		Avg: time.Duration(sum / int64(count)),
	}
}

type PipelineTimers struct {
	Decode        StageTimer
	Resize        StageTimer
	TemporalBlend StageTimer
	Scanline      StageTimer
	Quantize      StageTimer
	Dither        StageTimer
	Map           StageTimer
	Diff          StageTimer
	Output        StageTimer
	SyncWait      StageTimer
	FrameTotal    StageTimer
}
