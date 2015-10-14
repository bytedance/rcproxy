package proxy

import (
	"bufio"
	"net"
	"sync"

	log "github.com/ngaut/logging"
)

var (
	LogEveryN      uint64
	accessLogCount uint64
)

type Proxy struct {
	addr       string
	dispatcher *Dispatcher
	slotTable  *SlotTable
	connPool   *ConnPool
	exitChan   chan struct{}
}

func NewProxy(addr string, dispatcher *Dispatcher, connPool *ConnPool) *Proxy {
	p := &Proxy{
		addr:       addr,
		dispatcher: dispatcher,
		connPool:   connPool,
		exitChan:   make(chan struct{}),
	}
	return p
}

func (p *Proxy) Exit() {
	close(p.exitChan)
}

func (p *Proxy) handleConnection(cc net.Conn) {
	session := &Session{
		Conn:        cc,
		r:           bufio.NewReader(cc),
		backQ:       make(chan *PipelineResponse, 1000),
		closeSignal: &sync.WaitGroup{},
		reqWg:       &sync.WaitGroup{},
		connPool:    p.connPool,
		dispatcher:  p.dispatcher,
		rspHeap:     &PipelineResponseHeap{},
	}
	session.Run()
}

func (p *Proxy) Run() {
	tcpAddr, err := net.ResolveTCPAddr("tcp", p.addr)
	if err != nil {
		log.Fatal(err)
	}

	listener, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		log.Fatal(err)
	} else {
		log.Infof("proxy listens on %s", p.addr)
	}
	defer listener.Close()

	go p.dispatcher.Run()

	for {
		conn, err := listener.AcceptTCP()
		if err != nil {
			log.Error(err)
			continue
		}
		log.Infof("accept client: %s", conn.RemoteAddr())
		go p.handleConnection(conn)
	}
}
