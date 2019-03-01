# Benchmark for implementation of servers that support 1m connections


**Servers**：

1. 1_simple_tcp_server： a 1m-connections server implemented based on `goroutines per connection`
2. 2_epoll_server: a 1m-connections server implemented based on `epoll`
3. 3_epoll_server_throughputs: 	add throughputs and latency test for 2_epoll_server
4. 4_epoll_client: 	implement the client based on `epoll`
5. 5_multiple_client: 	use `multiple epoll` to manage connections in client
6. 6_multiple_server:  	use `multiple epoll` to manage connections in server
7. 7_server_prefork: 	use `prefork` style of apache to implement server
8. 8_server_workerpool: use `Reactor` pattern to implement multiple event loops
9. 9_few_clients_high_throughputs: a simple `goroutines per connection` server for test throughtputs and latency
10. 10_io_intensive_epoll_server: an io-bound `multiple epoll`  server
11. 11_io_intensive_goroutine:  an io-bound `goroutines per connection` server
12. 12_cpu_intensive_epoll_server: a cpu-bound `multiple epoll`  server
13. 13_cpu_intensive_goroutine  an cpu-bound `goroutines per connection` server
	
  

**中文介绍**:

1. [百万 Go TCP 连接的思考: epoll方式减少资源占用](https://colobu.com/2019/02/23/1m-go-tcp-connection/)
2. [百万 Go TCP 连接的思考2: 百万连接的服务器的性能](https://colobu.com/2019/02/27/1m-go-tcp-connection-2/)
3. [百万 Go TCP 连接的思考3: 低连接场景下的服务器的吞吐和延迟](https://colobu.com/2019/02/28/1m-go-tcp-connection-3/)
