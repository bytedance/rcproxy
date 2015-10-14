package proxy

import (
	"bufio"
	"bytes"
	"container/heap"
	"net"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/fatih/pool"

	"github.com/collinmsn/resp"
	log "github.com/ngaut/logging"
)

var (
	MOVED         = []byte("-MOVED")
	ASK           = []byte("-ASK")
	ASK_CMD_BYTES = []byte("+ASKING\r\n")
	BLACK_CMD_ERR = []byte("unsupported command")
)

type Session struct {
	net.Conn
	r           *bufio.Reader
	reqSeq      int64
	rspSeq      int64
	backQ       chan *PipelineResponse
	closed      bool
	closeSignal *sync.WaitGroup
	reqWg       *sync.WaitGroup
	rspHeap     *PipelineResponseHeap
	connPool    *ConnPool
	dispatcher  *Dispatcher
}

func (s *Session) Run() {
	s.closeSignal.Add(1)
	go s.WritingLoop()
	s.ReadingLoop()
}

// WritingLoop consumes backQ and send response to client
// It close the connection to notify reader on error
// and continue loop until the reader has exited
func (s *Session) WritingLoop() {
	for {
		select {
		case rsp, ok := <-s.backQ:
			if !ok {
				// reader has closed backQ
				s.Close()
				s.closeSignal.Done()
				return
			}

			if err := s.handleRespPipeline(rsp); err != nil {
				s.Close()
				continue
			}
		}
	}
}

func (s *Session) ReadingLoop() {
	for {
		cmd, err := resp.ReadCommand(s.r)
		if err != nil {
			log.Error(err)
			break
		}
		// convert all command name to upper case
		cmd.Args[0] = strings.ToUpper(cmd.Args[0])

		if n := atomic.AddUint64(&accessLogCount, 1); n%LogEveryN == 0 {
			if len(cmd.Args) > 1 {
				log.Infof("access %s %s %s", s.RemoteAddr(), cmd.Name(), cmd.Args[1])
			} else {
				log.Infof("access %s %s", s.RemoteAddr(), cmd.Name())
			}
		}

		// check if command is supported
		cmdFlag := CmdFlag(cmd)
		if cmdFlag&CMD_FLAG_BLACK != 0 {
			plReq := &PipelineRequest{
				seq: s.getNextReqSeq(),
				wg:  s.reqWg,
			}
			s.reqWg.Add(1)
			rsp := &resp.Data{T: resp.T_Error, String: BLACK_CMD_ERR}
			plRsp := &PipelineResponse{
				rsp: resp.NewObjectFromData(rsp),
				ctx: plReq,
			}
			s.backQ <- plRsp
			continue
		}

		// check if is multi key cmd
		if yes, numKeys := IsMultiCmd(cmd); yes && numKeys > 1 {
			s.handleMultiKeyCmd(cmd, numKeys)
			continue
		}

		// other general cmd
		key := cmd.Value(1)
		slot := Key2Slot(key)
		plReq := &PipelineRequest{
			cmd:      cmd,
			readOnly: cmdFlag&CMD_FLAG_READONLY != 0,
			slot:     slot,
			seq:      s.getNextReqSeq(),
			backQ:    s.backQ,
			wg:       s.reqWg,
		}

		s.reqWg.Add(1)
		s.dispatcher.Schedule(plReq)
	}
	// wait for all request done
	s.reqWg.Wait()

	// notify writer
	close(s.backQ)
	s.closeSignal.Wait()
}

// 将resp写出去。如果是multi key command，只有在全部完成后才汇总输出
func (s *Session) writeResp(plRsp *PipelineResponse) error {
	var buf []byte
	if parentCmd := plRsp.ctx.parentCmd; parentCmd != nil {
		// sub request
		parentCmd.OnSubCmdFinished(plRsp)
		if !parentCmd.Finished() {
			return nil
		}
		buf = parentCmd.CoalesceRsp().rsp.Raw()
		s.rspSeq++
	} else {
		buf = plRsp.rsp.Raw()
	}
	// write to client directly with non-buffered io
	if _, err := s.Write(buf); err != nil {
		log.Error(err)
		return err
	}

	return nil
}

// redirect send request to backend again to new server told by redis cluster
func (s *Session) redirect(server string, plRsp *PipelineResponse, ask bool) {
	var conn net.Conn
	var err error

	plRsp.err = nil
	conn, err = s.connPool.GetConn(server)
	if err != nil {
		log.Error(err)
		plRsp.err = err
		return
	}
	defer func() {
		if err != nil {
			log.Error(err)
			conn.(*pool.PoolConn).MarkUnusable()
		}
		conn.Close()
	}()

	reader := bufio.NewReader(conn)
	if ask {
		if _, err = conn.Write(ASK_CMD_BYTES); err != nil {
			plRsp.err = err
			return
		}
	}
	if _, err = conn.Write(plRsp.ctx.cmd.Format()); err != nil {
		plRsp.err = err
		return
	}
	if ask {
		if _, err = resp.ReadData(reader); err != nil {
			plRsp.err = err
			return
		}
	}
	obj := resp.NewObject()
	if err = resp.ReadDataBytes(reader, obj); err != nil {
		plRsp.err = err
	} else {
		plRsp.rsp = obj
	}
}

// handleResp handles MOVED and ASK redirection and call write response
func (s *Session) handleResp(plRsp *PipelineResponse) error {
	if plRsp.ctx.seq != s.rspSeq {
		panic("impossible")
	}
	plRsp.ctx.wg.Done()
	if plRsp.ctx.parentCmd == nil {
		s.rspSeq++
	}

	if plRsp.err != nil {
		s.dispatcher.TriggerReloadSlots()
		rsp := &resp.Data{T: resp.T_Error, String: []byte(plRsp.err.Error())}
		plRsp.rsp = resp.NewObjectFromData(rsp)
	} else {
		raw := plRsp.rsp.Raw()
		if raw[0] == resp.T_Error {
			if bytes.HasPrefix(raw, MOVED) {
				_, server := ParseRedirectInfo(string(raw))
				s.dispatcher.TriggerReloadSlots()
				s.redirect(server, plRsp, false)
			} else if bytes.HasPrefix(raw, ASK) {
				_, server := ParseRedirectInfo(string(raw))
				s.redirect(server, plRsp, true)
			}
		}
	}

	if plRsp.err != nil {
		return plRsp.err
	}

	if !s.closed {
		if err := s.writeResp(plRsp); err != nil {
			return err
		}
	}

	return nil
}

// handleRespPipeline handles the response if its sequence number is equal to session's
// response sequence number, otherwise, put it to a heap to keep the response order is same
// to request order
func (s *Session) handleRespPipeline(plRsp *PipelineResponse) error {
	if plRsp.ctx.seq != s.rspSeq {
		heap.Push(s.rspHeap, plRsp)
		return nil
	}

	if err := s.handleResp(plRsp); err != nil {
		return err
	}
	// continue to check the heap
	for {
		if rsp := s.rspHeap.Top(); rsp == nil || rsp.ctx.seq != s.rspSeq {
			return nil
		}
		rsp := heap.Pop(s.rspHeap).(*PipelineResponse)
		if err := s.handleResp(rsp); err != nil {
			return err
		}
	}
	return nil
}

func (s *Session) handleMultiKeyCmd(cmd *resp.Command, numKeys int) {
	var subCmd *resp.Command
	var err error
	mc := NewMultiKeyCmd(cmd, numKeys)
	// multi sub cmd share the same seq number
	seq := s.getNextReqSeq()
	readOnly := mc.CmdType() == MGET
	for i := 0; i < numKeys; i++ {
		switch mc.CmdType() {
		case MGET:
			subCmd, err = resp.NewCommand("GET", cmd.Value(i+1))
		case MSET:
			subCmd, err = resp.NewCommand("SET", cmd.Value(2*i+1), cmd.Value((2*i + 2)))
		case DEL:
			subCmd, err = resp.NewCommand("DEL", cmd.Value(i+1))
		}
		if err != nil {
			panic(err)
		}
		key := subCmd.Value(1)
		slot := Key2Slot(key)
		plReq := &PipelineRequest{
			cmd:       subCmd,
			readOnly:  readOnly,
			slot:      slot,
			seq:       seq,
			subSeq:    i,
			backQ:     s.backQ,
			parentCmd: mc,
			wg:        s.reqWg,
		}
		s.reqWg.Add(1)
		s.dispatcher.Schedule(plReq)
	}
}

func (s *Session) Close() {
	log.Infof("close session %p", s)
	if !s.closed {
		s.closed = true
		s.Conn.Close()
	}
}

func (s *Session) Read(p []byte) (int, error) {
	return s.r.Read(p)
}

func (s *Session) getNextReqSeq() (seq int64) {
	seq = s.reqSeq
	s.reqSeq++
	return
}

// ParseRedirectInfo parse slot redirect information from MOVED and ASK Error
func ParseRedirectInfo(msg string) (slot int, server string) {
	var err error
	parts := strings.Fields(msg)
	if len(parts) != 3 {
		log.Fatalf("invalid redirect message: %s", msg)
	}
	slot, err = strconv.Atoi(parts[1])
	if err != nil {
		log.Fatalf("invalid redirect message: %s", msg)
	}
	server = parts[2]
	return
}
