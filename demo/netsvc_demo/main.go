// demo/netsvc_demo/main.go
//
// 第4章网络服务演示：起两个本地服务（HTTP/TCP），并用客户端访问。
// 运行方式：go run ./demo/netsvc_demo
//
// 注意：会临时占用 :18080 和 :19000 两个端口（用了不常见端口避免冲突）。
package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

func main() {
	var wg sync.WaitGroup

	// 启动 HTTP 服务（独立 goroutine）
	wg.Add(1)
	go func() {
		defer wg.Done()
		runHTTPServer(":18080")
	}()

	// 启动 TCP 服务（独立 goroutine）
	wg.Add(1)
	go func() {
		defer wg.Done()
		runTCPServer(":19000")
	}()

	// 等服务启动
	time.Sleep(300 * time.Millisecond)

	// 客户端依次访问
	demoHTTPClient()
	demoTCPClient()

	fmt.Println("\n✅ 演示完成（按 Ctrl+C 退出长期服务）")
	wg.Wait()
}

// 3.1 HTTP 服务端
func runHTTPServer(addr string) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello from %s @ %s", addr, time.Now().Format("15:04:05"))
	})
	mux.HandleFunc("/api/echo", func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		defer r.Body.Close()
		w.Write(body)
	})
	http.ListenAndServe(addr, mux)
}

// 3.1b HTTP 客户端
func demoHTTPClient() {
	fmt.Println("\n=== 3.1 HTTP Client ===")
	client := &http.Client{Timeout: 3 * time.Second}

	// GET
	resp, err := client.Get("http://localhost:18080/")
	if err != nil {
		fmt.Println("GET 失败:", err)
		return
	}
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	fmt.Printf("GET / → HTTP %d, body: %s\n", resp.StatusCode, body)

	// POST
	resp2, _ := client.Post(
		"http://localhost:18080/api/echo",
		"text/plain",
		strings.NewReader("hello echo"),
	)
	body2, _ := io.ReadAll(resp2.Body)
	resp2.Body.Close()
	fmt.Printf("POST /api/echo → HTTP %d, body: %s\n", resp2.StatusCode, body2)
}

// 3.2 TCP 服务端
func runTCPServer(addr string) {
	ln, _ := net.Listen("tcp", addr)
	for {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		go func(c net.Conn) {
			defer c.Close()
			sc := bufio.NewScanner(c)
			for sc.Scan() {
				fmt.Fprintf(c, "ECHO: %s\n", sc.Text())
			}
		}(conn)
	}
}

// 3.2b TCP 客户端
func demoTCPClient() {
	fmt.Println("\n=== 3.2 TCP Client ===")
	conn, err := net.DialTimeout("tcp", "localhost:19000", 3*time.Second)
	if err != nil {
		fmt.Println("Dial 失败:", err)
		return
	}
	defer conn.Close()

	r := bufio.NewReader(conn)
	for _, msg := range []string{"hello", "world", "go!"} {
		fmt.Fprintln(conn, msg)
		resp, _ := r.ReadString('\n')
		fmt.Printf("发送 %q → 收到 %s", msg, resp)
	}
}
