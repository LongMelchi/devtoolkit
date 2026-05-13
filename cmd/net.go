// cmd/net.go
//
// net 子命令路由器（TCP 阶段：仅支持 tcp 协议）。
//
// 用法:
//
//	devtoolkit net -proto tcp -mode server -port 9000
//	devtoolkit net -proto tcp -mode client -host 127.0.0.1 -port 9000
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
	proto := fs.String("proto", "tcp", "协议: tcp") // ⚠️ 此版本只支持 tcp
	mode := fs.String("mode", "server", "模式: server|client")
	host := fs.String("host", "127.0.0.1", "客户端目标主机")
	port := fs.String("port", "9000", "端口")
	_ = fs.Parse(args)

	switch *proto {
	case "tcp":
		runTCP(*mode, *host, *port)
	default:
		fmt.Fprintf(os.Stderr, "❌ 不支持的协议: %s\n", *proto)
		os.Exit(config.ExitUsageError)
	}
}

// runTCP 处理 TCP 服务端/客户端。
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
