// main.go
// devtoolkit 主程序入口。
// 路由策略：
//  1. os.Args[1] 决定要进入哪个子命令分支
//  2. 各分支调用 cmd/* 中的处理函数
//  3. cmd/* 再调用 L1 业务包（filetool / netsvc / procman / ...）
//
// 启动流程：main → cmd.HandleXxx() → 业务包.RealLogic() → os.Exit(code)
package main

import (
	"fmt"
	"os"

	"devtoolkit/cmd"
	"devtoolkit/config"
)

func main() {
	// 第一步：加载配置（虽然这里没立刻用到，但保证子命令能拿到 cfg）
	cfg := config.Load()
	_ = cfg // 避免"声明但未使用"警告，后续章节会真正用上

	// 第二步：检查是否有子命令
	// len(os.Args) == 1 时，只有程序名，没有子命令
	if len(os.Args) < 2 {
		cmd.PrintRootHelp()
		os.Exit(config.ExitOK)
	}

	// 第三步：路由分发
	switch os.Args[1] {
	case "help", "-h", "--help":
		cmd.PrintRootHelp()
		os.Exit(config.ExitOK)
	case "version", "-v", "--version":
		cmd.PrintVersion()
		os.Exit(config.ExitOK)
	// === 第2章追加 ===
	case "file":
		cmd.HandleFile(os.Args[2:])

	// === 第3章追加 ===
	case "data":
		os.Exit(cmd.RunData(os.Args[2:]))

	// === 第4章追加 ===
	case "net":
		cmd.HandleNet(os.Args[2:])

	// === 第5章追加 ===
	// case "proc":
	//     cmd.HandleProc(os.Args[2:])

	// === 第6章追加：扩展子命令 ===
	// case "healthcheck": cmd.HandleHealthcheck(os.Args[2:])
	// case "backup":      cmd.HandleBackup(os.Args[2:])
	// case "procmon":     cmd.HandleProcmon(os.Args[2:])
	// case "config":      cmd.HandleConfig(os.Args[2:])
	// case "logagg":      cmd.HandleLogagg(os.Args[2:])
	// case "portscan":    cmd.HandlePortscan(os.Args[2:])
	default:
		cmd.UnknownCommand(os.Args[1])
		os.Exit(config.ExitUsageError)
	}
	fmt.Println() // 友好换行
}
