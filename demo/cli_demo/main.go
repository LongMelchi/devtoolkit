// demo/cli_demo/main.go
// 第1章所有语法点的统一演示程序。
// 运行方式：go run ./demo/cli_demo arg1 arg2 -flag value
// 提示：可单独注释掉某些 demoXxx() 调用只看感兴趣的部分。
package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
)

func main() {
	demoArgs()
	demoFlags()
	demoSubcommands()
	demoEnv()
}

// 1.1 os.Args 演示
func demoArgs() {
	fmt.Println("\n=== 1.1 os.Args 命令行参数 ===")
	fmt.Printf("程序路径 args[0]: %s\n", os.Args[0])
	fmt.Printf("参数个数 (不含路径): %d\n", len(os.Args)-1)
	for i, a := range os.Args[1:] {
		fmt.Printf("  args[%d] = %q\n", i+1, a)
	}
}

// 1.2 flag 包演示（注意：和 demoSubcommands 共用 os.Args，会冲突，只调用一处）
func demoFlags() {
	fmt.Println("\n=== 1.2 flag 标志解析 ===")

	// 注意：这里用 NewFlagSet 而非全局 flag，避免与外层 main 包的 flag 冲突
	// 创建一个名为"demo"的新命令行参数解析器 。
	// flag.ContinueOnError: 解析出错时返回错误，不退出程序
	fs := flag.NewFlagSet("demo", flag.ContinueOnError)

	host := fs.String("host", "localhost", "服务主机地址")
	port := fs.Int("port", 8080, "服务端口")
	// 布尔标志的特殊规则，出现"-debug"即为true
	debug := fs.Bool("debug", false, "是否开启调试")

	// 演示用：手工传入一组参数（避免读取真实 os.Args）
	demoArgs := []string{"-host", "0.0.0.0", "-port", "9090", "-debug"}
	if err := fs.Parse(demoArgs); err != nil {
		fmt.Println("解析失败:", err)
		return
	}

	fmt.Printf("host  = %s\n", *host)
	fmt.Printf("port  = %d\n", *port)
	fmt.Printf("debug = %t\n", *debug)
	fmt.Printf("剩余位置参数: %v\n", fs.Args())
}

// 1.3 子命令演示（NewFlagSet 实现 git/docker 风格的子命令）
func demoSubcommands() {
	fmt.Println("\n=== 1.3 子命令模式 ===")

	// 模拟用户输入: demo file -action read -file go.mod
	args := []string{"file", "-action", "read", "-file", "go.mod"}

	if len(args) < 1 {
		fmt.Println("未提供子命令")
		return
	}

	switch args[0] {
	case "file":
		fileCmd := flag.NewFlagSet("file", flag.ContinueOnError)
		action := fileCmd.String("action", "", "操作: read|write|list")
		filename := fileCmd.String("file", "", "文件名")
		_ = fileCmd.Parse(args[1:])
		fmt.Printf("✓ 子命令 file: action=%q, file=%q\n", *action, *filename)
	case "net":
		netCmd := flag.NewFlagSet("net", flag.ContinueOnError)
		host := netCmd.String("host", "localhost", "主机")
		_ = netCmd.Parse(args[1:])
		fmt.Printf("✓ 子命令 net: host=%q\n", *host)
	default:
		fmt.Printf("✗ 未知子命令: %s\n", args[0])
	}
}

// 1.4 环境变量演示
func demoEnv() {
	fmt.Println("\n=== 1.4 环境变量 ===")

	// Getenv: 不存在返回空字符串
	fmt.Printf("PATH 长度: %d\n", len(os.Getenv("PATH")))
	fmt.Printf("PATH: %v\n", os.Getenv("PATH"))

	// LookupEnv: 区分"未设置"和"空字符串"
	if v, ok := os.LookupEnv("DEVTOOLKIT_HOST"); ok {
		fmt.Printf("DEVTOOLKIT_HOST = %q (已设置)\n", v)
	} else {
		fmt.Println("DEVTOOLKIT_HOST 未设置")
	}

	// Setenv: 设置仅对当前进程有效
	os.Setenv("DEVTOOLKIT_DEMO", "hello")
	fmt.Printf("已设置 DEVTOOLKIT_DEMO = %q\n", os.Getenv("DEVTOOLKIT_DEMO"))

	// Environ: 返回所有环境变量
	all := os.Environ()
	fmt.Printf("当前进程共有 %d 个环境变量\n", len(all))
}

// 1.5 退出码演示（不真的调用 os.Exit，仅展示常量）
func demoExit() {
	fmt.Println("\n=== 1.5 退出码 ===")

	// 模拟一次"参数解析"
	demoUserInput := "ABC"
	if _, err := strconv.Atoi(demoUserInput); err != nil {
		fmt.Printf("解析失败 -> 应返回退出码 2 (config.ExitUsageError)\n")
		// 真正退出：os.Exit(2)（这里不调用，否则后续 demo 不会运行）
	}

	fmt.Println("提示: os.Exit(N) 会立刻终止程序，defer 不会执行！")
	fmt.Println("对比: return 会触发 defer，但只能从当前函数返回")
}
