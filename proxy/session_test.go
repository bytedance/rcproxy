package proxy

import (
	"container/heap"
	"errors"
	"testing"
)

var (
	sessionExpectedErr = errors.New("Session is expected")
)

// an ugly test for Session's handleRespPipeline
type sessionI interface {
	handleResp(plRsp *PipelineResponse) error
	handleRespPipeline(plRsp *PipelineResponse) error
}

type MockSession struct {
	sessionI
}

func (ms *MockSession) handleResp(plRsp *PipelineResponse) error {
	s, ok := ms.sessionI.(*Session)
	if !ok {
		panic(sessionExpectedErr)
	}
	if plRsp.ctx.seq != s.rspSeq {
		panic("request seq is not equal to response seq")
	}
	s.rspSeq++
	return nil
}

func (ms *MockSession) handleRespPipeline(plRsp *PipelineResponse) error {
	s, ok := ms.sessionI.(*Session)
	if !ok {
		panic(sessionExpectedErr)
	}
	if plRsp.ctx.seq != s.rspSeq {
		heap.Push(s.rspHeap, plRsp)
		return nil
	}

	if err := ms.handleResp(plRsp); err != nil {
		return err
	}
	// continue to check the heap
	for {
		if rsp := s.rspHeap.Top(); rsp == nil || rsp.ctx.seq != s.rspSeq {
			return nil
		}
		rsp := heap.Pop(s.rspHeap).(*PipelineResponse)
		if err := ms.handleResp(rsp); err != nil {
			return err
		}
	}
	return nil
}

func TestHandleRespPipeline(t *testing.T) {
	ms := &MockSession{
		sessionI: &Session{rspHeap: &PipelineResponseHeap{}},
	}
	seqs := []int{
		3, 7, 2, 1, 0, 4, 9, 5, 6, 8,
	}
	// response seq after each handleRespPipeline
	rspSeqs := []int{
		0, 0, 0, 0, 4, 5, 5, 6, 8, 10,
	}
	plRsps := []*PipelineResponse{}
	for _, seq := range seqs {
		plRsps = append(plRsps, &PipelineResponse{
			ctx: &PipelineRequest{
				seq: int64(seq),
			},
		})
	}
	for i, rsp := range plRsps {
		ms.handleRespPipeline(rsp)
		s, ok := ms.sessionI.(*Session)
		if !ok {
			panic(sessionExpectedErr)
		}
		if s.rspSeq != int64(rspSeqs[i]) {
			t.Errorf("expected rsp seq: %d, got %d", rspSeqs[i], s.rspSeq)
		}
	}
}
