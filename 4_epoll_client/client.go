package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"log"
	"net"
	"syscall"
	"time"

	"github.com/rcrowley/go-metrics"
)

var (
	ip          = flag.String("ip", "127.0.0.1", "server IP")
	connections = flag.Int("conn", 1, "number of tcp connections")
	startMetric = flag.String("sm", time.Now().Format("2006-01-02T15:04:05 -0700"), "start time point of all clients")
)

var (
	opsRate = metrics.NewRegisteredTimer("ops", nil)
)
var epoller *epoll

// client改造成epoll方式, 处理epoll消息是单线程的
func main() {
	flag.Parse()

	// setLimit()

	// go func() {
	// 	startPoint, err := time.Parse("2006-01-02T15:04:05 -0700", *startMetric)
	// 	if err != nil {
	// 		panic(err)
	// 	}
	// 	time.Sleep(startPoint.Sub(time.Now()))

	// 	metrics.Log(metrics.DefaultRegistry, 5*time.Second, log.New(os.Stderr, "metrics: ", log.Lmicroseconds))
	// }()

	var err error
	epoller, err = MkEpoll()
	if err != nil {
		panic(err)
	}

	addr := *ip + ":8972"
	log.Printf("连接到 %s", addr)
	var conns []net.Conn
	for i := 0; i < *connections; i++ {
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

	tts := time.Second
	if *connections > 100 {
		tts = time.Millisecond * 5
	}

	go start()

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

func start() {
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
				log.Printf("failed to read %v", err)
				if err := epoller.Remove(conn); err != nil {
					log.Printf("failed to remove %v", err)
				}

				conn.Close()
				continue
			} else {
				opsRate.Update(time.Duration(time.Now().UnixNano() - nano))
			}

			err = binary.Write(conn, binary.BigEndian, time.Now().UnixNano())
			if err != nil {
				log.Printf("failed to write %v", err)
				if err := epoller.Remove(conn); err != nil {
					log.Printf("failed to remove %v", err)
				}
				conn.Close()
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
