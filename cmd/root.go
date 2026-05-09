// 设计说明：根命令负责打印帮助、版本信息、提示可用子命令。
// 它不依赖任何 L1 模块，是路由层的"门面"。

// cmd/root.go
// 根命令：当用户运行 `devtoolkit` 不带任何子命令时显示帮助。
// 当用户运行 `devtoolkit help` 或 `devtoolkit -h` 时也显示这里。
package cmd

import (
	"fmt"
	"os"
)

// Version 是项目版本号，发布时通过 -ldflags 在编译期注入。
// 编译指令示例：
//
//	go build -ldflags "-X devtoolkit/cmd.Version=v1.0.0" -o devtoolkit
var Version = "dev"

// PrintRootHelp 打印工具箱的总览帮助。
// 子命令（file/net/proc/...）会陆续在后面章节注册到这里。
func PrintRootHelp() {
	fmt.Println(`devtoolkit - 简易 DevOps 工具箱

用法:
  devtoolkit <命令> [选项]

核心命令:
  file         文件操作（读/写/搜索/遍历）
  net          网络服务（HTTP/TCP 客户端与服务端）
  proc         进程管理（执行命令/日志）

数据命令（第3章加入）:
  data         运行数据本地存储查询（list/tail/clear）

扩展命令（第6章逐步加入）:
  healthcheck  服务健康检查
  backup       文件备份
  procmon      进程监控
  config       配置中心
  logagg       日志聚合
  portscan     端口扫描

通用命令:
  help         显示本帮助
  version      显示版本号

环境变量:
  DEVTOOLKIT_HOST       默认主机（localhost）
  DEVTOOLKIT_PORT       默认端口（8080）
  DEVTOOLKIT_LOG        日志级别（info）
  DEVTOOLKIT_LOGFILE    日志文件路径（空=仅控制台）
  DEVTOOLKIT_DATA_DIR   运行数据目录（默认 ./data）

示例:
  devtoolkit file -action read -file go.mod
  devtoolkit data list
  devtoolkit net -proto http -mode server -port 9090
  devtoolkit proc -exec "go version"`)
}

// PrintVersion 打印版本号到 stdout。
func PrintVersion() {
	fmt.Printf("devtoolkit %s\n", Version)
}

// UnknownCommand 处理未识别的子命令。
// 写到 stderr（os.Stderr）而非 stdout，是 Unix 工具的最佳实践：
// 错误信息走 stderr，正常输出走 stdout，便于 shell 中分流。
func UnknownCommand(name string) {
	fmt.Fprintf(os.Stderr, "❌ 未知命令: %s\n\n", name)
	PrintRootHelp()
}
