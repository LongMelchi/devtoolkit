// netsvc/http_server.go
//
// HTTP 服务端封装：把 net/http 的常用 pattern 包装成易用的 API。
//
// 设计：
//   - 不直接调用 http.HandleFunc 全局函数（会污染 DefaultServeMux）
//   - 用独立的 *http.ServeMux 实例，便于测试和并发隔离
//
// // 方式1：使用全局函数（不推荐）
// http.HandleFunc("/api/users", handleUsers)  // 注册到全局 DefaultServeMux
//
// // 方式2：使用独立实例（推荐）
// mux := http.NewServeMux()
// mux.HandleFunc("/api/users", handleUsers)   // 注册到独立 mux
package netsvc

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// HTTPServer 是 HTTP 服务端的轻量封装。
type HTTPServer struct {
	addr   string         // 监听地址（如 ":8080"）
	mux    *http.ServeMux // 路由复用器
	server *http.Server   // 底层 http.Server，用于优雅关闭
}

// NewHTTPServer 创建一个 HTTPServer，并注册 3 个内置端点。
//
// 内置端点：
//
//	GET  /         返回欢迎信息
//	GET  /api/info JSON 格式的服务元数据
//	POST /api/echo 把请求体原样回显
func NewHTTPServer(addr string) *HTTPServer {
	mux := http.NewServeMux()
	s := &HTTPServer{
		addr: addr,
		mux:  mux,
		server: &http.Server{
			Addr:         addr,
			Handler:      mux,
			ReadTimeout:  10 * time.Second, // 防止慢客户端 DoS
			WriteTimeout: 10 * time.Second,
			IdleTimeout:  60 * time.Second,
		},
	}
	s.registerBuiltinHandlers()
	return s
}

// registerBuiltinHandlers 注册内置路由。
//
// 私有函数，外部不可调用，确保所有 HTTPServer 实例都有这些端点。
func (s *HTTPServer) registerBuiltinHandlers() {
	s.mux.HandleFunc("/", s.handleRoot)
	s.mux.HandleFunc("/api/info", s.handleInfo)
	s.mux.HandleFunc("/api/echo", s.handleEcho)
}

// handleRoot 处理 / 根路径。 w是写向客户端的，r是有服务端读取的
func (s *HTTPServer) handleRoot(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "devtoolkit HTTP Server\n")
	fmt.Fprintf(w, "Method: %s\n", r.Method)
	fmt.Fprintf(w, "Path:   %s\n", r.URL.Path)
	fmt.Fprintf(w, "Time:   %s\n", time.Now().Format(time.RFC3339))
}

// handleInfo 返回 JSON 格式的服务信息。
//
// 注意先设置 Content-Type，再写 body；顺序反过来则 header 已发送，无效。
func (s *HTTPServer) handleInfo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	// json.NewEncoder(w): 创建一个 JSON 编码器，目标写入 w （HTTP 响应）
	json.NewEncoder(w).Encode(map[string]any{
		"name":    "devtoolkit",
		"version": "1.0",
		"time":    time.Now().Format(time.RFC3339),
	})
}

// handleEcho 把请求体原样写回响应（用于客户端测试）。
func (s *HTTPServer) handleEcho(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close() // 一定要关，否则连接泄漏
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "读取请求体失败: "+err.Error(), http.StatusBadRequest)
		return
	}
	w.Write(body)
}

// Start 启动 HTTP 服务，阻塞当前 goroutine。
//
// 通常用法：在 main 中 go server.Start(); 然后处理信号；最后 server.Shutdown()。
func (s *HTTPServer) Start() error {
	fmt.Printf("🌐 HTTP 服务启动于 http://localhost%s\n", s.addr)
	fmt.Println("   端点: /  /api/info  /api/echo")
	if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("HTTP 服务启动失败: %w", err)
	}
	return nil
}

// Shutdown 优雅关闭：等待进行中的请求完成，最多等 timeout。
//
// 配合 signal 实现 Ctrl+C 优雅退出。第6章会用到。
func (s *HTTPServer) Shutdown(timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return s.server.Shutdown(ctx)
}
