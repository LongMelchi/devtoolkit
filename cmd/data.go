// 设计说明：`data` 子命令对外暴露 `list / tail / clear` 三个 action。
// cmd/data.go
//
// data 子命令路由器。
// 用户调用：
//
//	devtoolkit data list                 列出所有 .jsonl 文件
//	devtoolkit data tail -module file    打印 file 模块最新事件
//	devtoolkit data tail -module file -n 5
//	devtoolkit data clear -module file   清空 file 模块的当天文件
package cmd

import (
	"flag"
	"fmt"
	"os"

	"devtoolkit/config"
	"devtoolkit/datastore"
)

// RunData 是 data 子命令的入口，被 main.go 调用。
//
// args 已是去掉子命令名后的参数（如 ["tail", "-module", "file"]）。
func RunData(args []string) int {
	if len(args) == 0 {
		printDataHelp()
		return config.ExitUsageError
	}

	action := args[0]
	rest := args[1:]
	cfg := config.Load()

	switch action {
	case "list":
		return runDataList(cfg.DataDir, rest)
	case "tail":
		return runDataTail(cfg.DataDir, rest)
	case "clear":
		return runDataClear(cfg.DataDir, rest)
	case "help", "-h", "--help":
		printDataHelp()
		return config.ExitOK
	default:
		fmt.Fprintf(os.Stderr, "未知的 data 子动作: %s\n", action)
		printDataHelp()
		return config.ExitUsageError
	}
}

func runDataList(dir string, args []string) int {
	fs := flag.NewFlagSet("data list", flag.ContinueOnError)
	module := fs.String("module", "", "仅列出指定模块的文件（空 = 全部）")
	if err := fs.Parse(args); err != nil {
		return config.ExitUsageError
	}

	files, err := datastore.ListFiles(dir, *module)
	if err != nil {
		fmt.Fprintf(os.Stderr, "list 失败: %v\n", err)
		return config.ExitIOError
	}
	if len(files) == 0 {
		fmt.Println("(空) 还没有任何运行数据")
		return config.ExitOK
	}
	for _, f := range files {
		info, _ := os.Stat(f)
		size := int64(0)
		if info != nil {
			size = info.Size()
		}
		fmt.Printf("  %s  (%d bytes)\n", f, size)
	}
	return config.ExitOK
}

func runDataTail(dir string, args []string) int {
	fs := flag.NewFlagSet("data tail", flag.ContinueOnError)
	module := fs.String("module", "", "模块名（必填）")
	n := fs.Int("n", 10, "显示最近 N 条")
	if err := fs.Parse(args); err != nil {
		return config.ExitUsageError
	}
	if *module == "" {
		fmt.Fprintln(os.Stderr, "tail 必须指定 -module")
		return config.ExitUsageError
	}

	files, err := datastore.ListFiles(dir, *module)
	if err != nil || len(files) == 0 {
		fmt.Fprintf(os.Stderr, "找不到模块 %s 的数据\n", *module)
		return config.ExitIOError
	}

	// 取最新一个文件 (ListFiles 已倒序)
	events, err := datastore.ReadAll(files[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "读取失败: %v\n", err)
		return config.ExitIOError
	}

	start := 0
	if len(events) > *n {
		start = len(events) - *n
	}
	for _, e := range events[start:] {
		fmt.Printf("[%s] %s.%s status=%s msg=%q\n",
			e.Time.Format("15:04:05"), e.Module, e.Action, e.Status, e.Message)
	}
	return config.ExitOK
}

func runDataClear(dir string, args []string) int {
	fs := flag.NewFlagSet("data clear", flag.ContinueOnError)
	module := fs.String("module", "", "模块名（必填）")
	if err := fs.Parse(args); err != nil {
		return config.ExitUsageError
	}
	if *module == "" {
		fmt.Fprintln(os.Stderr, "clear 必须指定 -module")
		return config.ExitUsageError
	}

	files, _ := datastore.ListFiles(dir, *module)
	for _, f := range files {
		if err := os.Remove(f); err != nil {
			fmt.Fprintf(os.Stderr, "删除 %s 失败: %v\n", f, err)
			return config.ExitIOError
		}
		fmt.Printf("已删除: %s\n", f)
	}
	if len(files) == 0 {
		fmt.Println("(无文件可删)")
	}
	return config.ExitOK
}

func printDataHelp() {
	fmt.Println(`data - 运行数据本地存储查询工具

用法:
  devtoolkit data <动作> [选项]

动作:
  list                    列出所有运行数据文件
  tail  -module <name>    查看模块最近事件（默认最后 10 条）
        [-n N]
  clear -module <name>    清空模块当天的运行数据

环境变量:
  DEVTOOLKIT_DATA_DIR     数据存放目录（默认 ./data）

示例:
  devtoolkit data list
  devtoolkit data list  -module file
  devtoolkit data tail  -module file -n 5
  devtoolkit data clear -module file`)
}
