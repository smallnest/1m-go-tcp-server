package main

import (
	"io"
	"log"
	"net"
	"sync"
)

type pool struct {
	workers   int
	maxTasks  int
	taskQueue chan net.Conn

	mu     sync.Mutex
	closed bool
	done   chan struct{}
}

func newPool(w int, t int) *pool {
	return &pool{
		workers:   w,
		maxTasks:  t,
		taskQueue: make(chan net.Conn, t),
		done:      make(chan struct{}),
	}
}

func (p *pool) Close() {
	p.mu.Lock()
	p.closed = true
	close(p.done)
	close(p.taskQueue)
	p.mu.Unlock()
}

func (p *pool) addTask(conn net.Conn) {
	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		return
	}
	p.mu.Unlock()

	p.taskQueue <- conn
}

func (p *pool) start() {
	for i := 0; i < p.workers; i++ {
		go p.startWorker()
	}
}

func (p *pool) startWorker() {
	for {
		select {
		case <-p.done:
			return
		case conn := <-p.taskQueue:
			if conn != nil {
				handleConn(conn)
			}
		}
	}
}

func handleConn(conn net.Conn) {
	_, err := io.CopyN(conn, conn, 8)
	if err != nil {
		if err := epoller.Remove(conn); err != nil {
			log.Printf("failed to remove %v", err)
		}
		conn.Close()
	}
	opsRate.Mark(1)
}
