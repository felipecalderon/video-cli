package pipeline

import "video-terminal/types"

// Slot representa un conjunto de buffers pre-reservados para un frame completo.
// Viaja a través del pipeline concurrente para evitar allocations.
type Slot struct {
	Frame    types.FrameRGB
	Work     types.WorkRGB
	WorkTemp types.WorkRGB // Para el double-buffering de Temporal Blend
	Grid     types.CellGrid
}

type SlotPool struct {
	slots chan *Slot
}

// NewSlotPool inicializa un pool de tamaño fijo.
func NewSlotPool(size int) *SlotPool {
	p := &SlotPool{
		slots: make(chan *Slot, size),
	}
	for i := 0; i < size; i++ {
		p.slots <- &Slot{}
	}
	return p
}

func (p *SlotPool) Acquire() *Slot {
	return <-p.slots
}

func (p *SlotPool) Release(s *Slot) {
	p.slots <- s
}
