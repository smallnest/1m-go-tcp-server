//+build !linux

package main

import (
	"net"
	"sync"
)

type epoll struct {
	fd          int
	connections map[int]net.Conn
	lock        *sync.RWMutex
}

func MkEpoll() (*epoll, error) {

	return nil, nil
}

func (e *epoll) Add(conn net.Conn) error {
	return nil
}

func (e *epoll) Remove(conn net.Conn) error {
	return nil
}

func (e *epoll) Wait() ([]net.Conn, error) {

	return nil, nil
}
