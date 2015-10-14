package proxy

import (
	"sync"

	"github.com/collinmsn/resp"
)

type PipelineRequest struct {
	cmd *resp.Command
	// if it is readOnly command
	readOnly bool
	// key slot
	slot int
	// session wide request sequence number
	seq int64
	// sub sequence number for multi key command
	subSeq int
	backQ  chan *PipelineResponse
	// session wide wait group
	wg *sync.WaitGroup
	// for multi key command, owner of this command
	parentCmd *MultiKeyCmd
}

type PipelineResponse struct {
	rsp *resp.Object
	ctx *PipelineRequest
	err error
}

type PipelineResponseHeap []*PipelineResponse

func (h PipelineResponseHeap) Len() int {
	return len(h)
}
func (h PipelineResponseHeap) Less(i, j int) bool {
	return h[i].ctx.seq < h[j].ctx.seq
}
func (h PipelineResponseHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
}
func (h *PipelineResponseHeap) Push(x interface{}) {
	*h = append(*h, x.(*PipelineResponse))
}
func (h *PipelineResponseHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

// Peek will return the heap top element
func (h PipelineResponseHeap) Top() *PipelineResponse {
	if h.Len() == 0 {
		return nil
	}
	return h[0]
}
