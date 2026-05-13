// cmd/net.go
//
// net 子命令路由器（HTTP + TCP 合并版：同时支持两种协议）。
//
// 用法:
//
//	devtoolkit net -proto http -mode server -port 8080
//	devtoolkit net -proto http -mode client -url http://localhost:8080/api/info
//	devtoolkit net -proto tcp  -mode server -port 9000
//	devtoolkit net -proto tcp  -mode client -host 127.0.0.1 -port 9000
package cmd

import (
	"bufio"
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
	proto := fs.String("proto", "http", "协议: http|tcp")
	mode := fs.String("mode", "server", "模式: server|client")
	host := fs.String("host", "127.0.0.1", "客户端目标主机（tcp/http 皆用）")
	port := fs.String("port", "8080", "端口")
	url := fs.String("url", "http://localhost:8080/api/info", "HTTP 客户端目标 URL")
	_ = fs.Parse(args)

	switch *proto {
	case "http":
		runHTTP(*mode, *port, *url)
	case "tcp":
		runTCP(*mode, *host, *port)
	default:
		fmt.Fprintf(os.Stderr, "❌ 不支持的协议: %s\n", *proto)
		os.Exit(config.ExitUsageError)
	}
}

// runHTTP 处理 HTTP 服务端/客户端（HTTP 阶段实现）。
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

// runTCP 处理 TCP 服务端/客户端（TCP 阶段实现）。
func runTCP(mode, host, port string) {
	switch mode {
	case "server":
		s := netsvc.NewTCPServer(":" + port)
		if err := s.Start(); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(config.ExitNetworkError)
		}
	case "client":
		c, err := netsvc.DialTCP(host, port, 5*time.Second)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(config.ExitNetworkError)
		}
		defer c.Close()
		fmt.Println("已连接，输入消息（quit 退出）:")
		stdin := bufio.NewScanner(os.Stdin)
		for stdin.Scan() {
			resp, err := c.SendLine(stdin.Text())
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				return
			}
			fmt.Print(resp)
			if stdin.Text() == "quit" {
				return
			}
		}
	default:
		fmt.Fprintf(os.Stderr, "未知模式: %s\n", mode)
		os.Exit(config.ExitUsageError)
	}
}
