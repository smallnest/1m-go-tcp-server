package main

import (
	"encoding/binary"
	"flag"
	"fmt"
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
	startMetric = flag.String("sm", time.Now().Format("2006-01-02T15:04:05 -0700"), "start time point of all clients")
)

var (
	opsRate = metrics.NewRegisteredTimer("ops", nil)
)

// client采用epoll方式，但是会多线程的处理
func main() {
	flag.Parse()

	//setLimit()
	go func() {
		startPoint, _ := time.Parse("2006-01-02T15:04:05 -0700", *startMetric)
		time.Sleep(startPoint.Sub(time.Now()))

		metrics.Log(metrics.DefaultRegistry, 5*time.Second, log.New(os.Stderr, "metrics: ", log.Lmicroseconds))
	}()

	addr := *ip + ":8972"
	log.Printf("连接到 %s", addr)

	for i := 0; i < *c; i++ {
		go mkClient(addr, *connections/(*c))
	}

	select {}
}

func mkClient(addr string, connections int) {
	epoller, err := MkEpoll()
	if err != nil {
		panic(err)
	}

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

	tts := time.Second
	if *c > 100 {
		tts = time.Millisecond * 5
	}

	for i := 0; i < len(conns); i++ {
		time.Sleep(tts)
		conn := conns[i]
		err = binary.Write(conn, binary.BigEndian, time.Now().UnixNano())
		if err != nil {
			log.Printf("failed to write timestamp %v", err)
			if err := epoller.Remove(conn); err != nil {
				if err := epoller.Remove(conn); err != nil {
					log.Printf("failed to remove %v", err)
				}
			}
		}
	}

	select {}
}

func start(epoller *epoll) {
	var nano int64
	for {
		connections, err := epoller.Wait()
		if err != nil {
			log.Printf("failed to epoll wait %v", err)
			continue
		}
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

			err = binary.Write(conn, binary.BigEndian, time.Now().UnixNano())
			if err != nil {
				if err := epoller.Remove(conn); err != nil {
					log.Printf("failed to remove %v", err)
				}
			}
		}
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
