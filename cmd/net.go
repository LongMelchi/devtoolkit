// cmd/net.go
//
// net 子命令路由器（HTTP 阶段：仅支持 http 协议）。
//
// 用法:
//
//	devtoolkit net -proto http -mode server -port 8080
//	devtoolkit net -proto http -mode client -url http://localhost:8080/api/info
package cmd

import (
	"flag"
	"fmt"
	"os"
	"time"

	"devtoolkit/config"
	"devtoolkit/netsvc"
)

// HandleNet 是 net 子命令的入口。
func HandleNet(args []string) {
	fs := flag.NewFlagSet("net", flag.ExitOnError)
	proto := fs.String("proto", "http", "协议: http") // ⚠️ 此版本只支持 http
	mode := fs.String("mode", "server", "模式: server|client")
	port := fs.String("port", "8080", "端口")
	url := fs.String("url", "http://localhost:8080/api/info", "客户端目标 URL")
	_ = fs.Parse(args)

	switch *proto {
	case "http":
		runHTTP(*mode, *port, *url)
	default:
		fmt.Fprintf(os.Stderr, "❌ 不支持的协议: %s\n", *proto)
		os.Exit(config.ExitUsageError)
	}
}

// runHTTP 处理 HTTP 服务端/客户端。
func runHTTP(mode, port, url string) {
	switch mode {
	case "server":
		s := netsvc.NewHTTPServer(":" + port)
		if err := s.Start(); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(config.ExitNetworkError)
		}
	case "client":
		c := netsvc.NewHTTPClient(5 * time.Second)
		code, body, err := c.Get(url)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(config.ExitNetworkError)
		}
		fmt.Printf("HTTP %d\n%s", code, string(body))
	default:
		fmt.Fprintf(os.Stderr, "未知模式: %s\n", mode)
		os.Exit(config.ExitUsageError)
	}

}
