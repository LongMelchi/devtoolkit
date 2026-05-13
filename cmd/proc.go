// cmd/proc.go
//
// proc 子命令路由器。
//
// 用法:
//
//	devtoolkit proc -exec "go version"          (同步执行)
//	devtoolkit proc -spawn "ping -c 3 1.1.1.1"  (派生执行，看 PID)
//	devtoolkit proc -log app.log -msg "hello"   (写一条日志)
package cmd

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"devtoolkit/config"
	"devtoolkit/procman"
)

// HandleProc 是 proc 子命令的入口。
func HandleProc(args []string) {
	fs := flag.NewFlagSet("proc", flag.ExitOnError)
	exec := fs.String("exec", "", "同步执行命令字符串，如 \"go version\"")
	spawn := fs.String("spawn", "", "派生执行命令字符串（拿到 PID 后立即返回）")
	logFile := fs.String("log", "", "日志文件路径（与 -msg 配合）")
	msg := fs.String("msg", "", "要写入日志的内容")
	_ = fs.Parse(args)

	switch {
	case *exec != "":
		parts := strings.Fields(*exec)
		if len(parts) == 0 {
			fmt.Fprintln(os.Stderr, "❌ -exec 参数为空")
			os.Exit(config.ExitUsageError)
		}
		result := procman.Run(parts[0], parts[1:]...)
		fmt.Print(result.Stdout)
		if result.Stderr != "" {
			fmt.Fprint(os.Stderr, result.Stderr)
		}
		fmt.Fprintf(os.Stderr, "\n[exit=%d, took=%v]\n", result.ExitCode, result.Duration)
		os.Exit(result.ExitCode)

	case *spawn != "":
		parts := strings.Fields(*spawn)
		p, err := procman.Spawn(nil, os.Stdout, os.Stderr, parts[0], parts[1:]...)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(config.ExitProcessError)
		}
		fmt.Printf("👶 派生子进程 PID=%d, 命令=%q\n", p.PID, p.Command)
		code, _ := p.Wait()
		fmt.Printf("✅ 子进程已退出（code=%d）\n", code)

	case *msg != "":
		lg, err := procman.NewLogger("[devtool] ", *logFile)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(config.ExitIOError)
		}
		defer lg.Close()
		lg.Println(*msg)
	default:
		fmt.Fprintln(os.Stderr,
			"用法: proc [-exec \"cmd args\"] | [-spawn \"cmd args\"] | [-log file -msg text]")
		os.Exit(config.ExitUsageError)
	}

}
