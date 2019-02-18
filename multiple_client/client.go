package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"syscall"
	"time"

	"github.com/rcrowley/go-metrics"
)

var (
	ip          = flag.String("ip", "127.0.0.1", "server IP")
	connections = flag.Int("conn", 1, "number of total tcp connections")
	c           = flag.Int("c", 100, "currency count")
)

var (
	opsRate = metrics.NewRegisteredTimer("ops", nil)
)
var epoller *epoll

func main() {
	flag.Usage = func() {
		io.WriteString(os.Stderr, `tcp客户端测试工具
使用方法: ./client -ip=172.17.0.1 -conn=10
`)
		flag.PrintDefaults()
	}
	flag.Parse()

	setLimit()
	go metrics.Log(metrics.DefaultRegistry, 5*time.Second, log.New(os.Stderr, "metrics: ", log.Lmicroseconds))

	var err error
	epoller, err = MkEpoll()
	if err != nil {
		panic(err)
	}

	addr := *ip + ":8972"
	log.Printf("连接到 %s", addr)

	for i := 0; i < *c; i++ {
		mkConn(addr, *connections/(*c))
	}

	select {}
}

func mkConn(addr string, connections int) {
	var conns []net.Conn
	for i := 0; i < connections; i++ {
		c, err := net.DialTimeout("tcp", addr, 10*time.Second)
		if err != nil {
			fmt.Println("failed to connect", i, err)
			i--
			continue
		}
		if err := epoller.Add(c); err != nil {
			log.Printf("failed to add connection %v", err)
			c.Close()
		}
		conns = append(conns, c)
	}

	log.Printf("完成初始化 %d 连接", len(conns))

	go start(epoller)

	for i := 0; i < len(conns); i++ {
		conn := conns[i]
		conn.Write([]byte("hello world\r\n"))
	}
}

func start(epoller *epoll) {
	var nano int64
	for {
		connections, err := epoller.Wait()
		if err != nil {
			log.Printf("failed to epoll wait %v", err)
			continue
		}
		go func() {
			for _, conn := range connections {
				if conn == nil {
					break
				}

				if err := binary.Read(conn, binary.BigEndian, &nano); err != nil {
					if err := epoller.Remove(conn); err != nil {
						log.Printf("failed to remove %v", err)
					}
				} else {
					opsRate.Update(time.Duration(time.Now().UnixNano() - nano))
				}

				err = binary.Write(conn, binary.BigEndian, []byte("hello world\r\n"))
				if err != nil {
					if err := epoller.Remove(conn); err != nil {
						log.Printf("failed to remove %v", err)
					}
				}
			}
		}()
	}
}

func setLimit() {
	var rLimit syscall.Rlimit
	if err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit); err != nil {
		panic(err)
	}
	rLimit.Cur = rLimit.Max
	if err := syscall.Setrlimit(syscall.RLIMIT_NOFILE, &rLimit); err != nil {
		panic(err)
	}
}
