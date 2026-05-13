// netsvc/tcp_server.go
//
// TCP 服务端：每个连接一个 goroutine 处理（最经典的并发模型）。
//
// 设计：把"接受连接"的循环和"处理单连接"的逻辑分开，便于测试和扩展。
package netsvc

import (
	"bufio"
	"fmt"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

// TCPServer 是支持优雅关闭的 echo 服务端。
type TCPServer struct {
	addr     string
	listener net.Listener
	wg       sync.WaitGroup // 等待所有连接处理完
	closing  atomic.Bool    // 标记是否正在关闭
}

// NewTCPServer 创建 TCP 服务端实例。
func NewTCPServer(addr string) *TCPServer {
	return &TCPServer{addr: addr}
}

// Start 启动监听并阻塞循环接受连接。
//
// 每个新连接派发到独立的 goroutine 执行 handleConn。
// 这是 Go 并发的精髓：开 goroutine 几乎免费（约 2KB 栈），
// 不像传统语言开线程那么贵。
func (s *TCPServer) Start() error {
	ln, err := net.Listen("tcp", s.addr)
	if err != nil {
		return fmt.Errorf("监听 %q 失败: %w", s.addr, err)
	}
	s.listener = ln
	fmt.Printf("🔌 TCP 服务启动于 %s\n", s.addr)

	for {
		conn, err := ln.Accept()
		if err != nil {
			if s.closing.Load() {
				return nil // 正常关闭
			}
			fmt.Println("接受连接失败:", err)
			continue
		}
		s.wg.Add(1)
		go s.handleConn(conn) // 关键：每个连接一个 goroutine
	}
}

// handleConn 处理单个连接：echo + 行回显。
//
// 这里使用 bufio.Scanner 按行接收，方便文本协议（telnet 测试也能用）。
func (s *TCPServer) handleConn(conn net.Conn) {
	defer s.wg.Done()
	defer conn.Close()

	addr := conn.RemoteAddr().String()
	fmt.Printf("🆕 新连接: %s\n", addr)
	defer fmt.Printf("👋 关闭连接: %s\n", addr)

	// 设置读超时，防止恶意客户端"挂着不发数据"
	conn.SetReadDeadline(time.Now().Add(5 * time.Minute))

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		line := scanner.Text()
		fmt.Printf("[%s] 收到: %q\n", addr, line)

		// 服务端回显：原样 + ECHO 前缀
		fmt.Fprintf(conn, "ECHO: %s\n", line)

		// 收到 quit 主动断开
		if line == "quit" {
			fmt.Fprintln(conn, "Bye!")
			return
		}
	}
}

// Shutdown 优雅关闭：先停止接受新连接，再等已有连接处理完。
func (s *TCPServer) Shutdown(timeout time.Duration) error {
	s.closing.Store(true)
	if s.listener != nil {
		s.listener.Close()
	}

	// 用 channel + select 实现"超时等待"
	done := make(chan struct{})
	go func() { s.wg.Wait(); close(done) }()
	select {
	case <-done:
		return nil
	case <-time.After(timeout):
		return fmt.Errorf("等待连接关闭超时（%v）", timeout)
	}
}
