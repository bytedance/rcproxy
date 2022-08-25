package proxy

import (
	"container/heap"
	"testing"
)

func TestPipelineResponseHeap(t *testing.T) {
	h := PipelineResponseHeap{}
	if h.Len() != 0 {
		t.Error("init length should be 0")
	}
	N := 10
	for i := 0; i < N; i++ {
		req := &PipelineRequest{
			seq: int64(i),
		}
		rsp := &PipelineResponse{
			ctx: req,
		}
		heap.Push(&h, rsp)
		if h.Len() != i+1 {
			t.Errorf("expected len: %d, got: %d", i+1, h.Len())
		}
	}
	for i := 0; i < N; i++ {
		min := h.Top()
		if min.ctx.seq != int64(i) {
			t.Errorf("expected heap min: %d, got: %d", i, min.ctx.seq)
		}
		heap.Pop(&h)
	}
}
