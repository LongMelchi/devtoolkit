// 作用：把 `file` 子命令的命令行参数解析与 `filetool/` 模块的调用绑定起来。
// cmd/file.go
//
// file 子命令路由器。
//
// 用法示例:
//
//	devtoolkit file -action read    -file go.mod
//	devtoolkit file -action grep    -file app.log -pattern ERROR
//	devtoolkit file -action list    -file logs/
//	devtoolkit file -action embed   (列出嵌入资源)
//	devtoolkit file -action help    (打印嵌入的 help.txt)
package cmd

import (
	"flag"
	"fmt"
	"os"

	"devtoolkit/config"
	"devtoolkit/filetool"
)

// HandleFile 是 file 子命令的入口，由 main.go 调用。
// 入参 args 是已经剥掉 "file" 之后的剩余参数（即 os.Args[2:]）。
func HandleFile(args []string) {
	fs := flag.NewFlagSet("file", flag.ExitOnError)
	action := fs.String("action", "", "操作: read|write|append|list|grep|embed|help")
	file := fs.String("file", "", "文件路径")
	pattern := fs.String("pattern", "", "grep 关键字（支持正则）")
	data := fs.String("data", "", "write/append 操作的写入内容")
	regex := fs.Bool("regex", false, "grep 时是否启用正则")

	// 解析 args，将解析出的值绑定到上面定义的变量
	_ = fs.Parse(args)

	switch *action {
	case "read":
		content, err := filetool.ReadAllString(*file)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(config.ExitIOError)
		}
		fmt.Print(content)

	case "write":
		if err := filetool.WriteFile(*file, []byte(*data)); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(config.ExitIOError)
		}
		fmt.Printf("✓ 已写入 %s (%d 字节)\n", *file, len(*data))

	case "append":
		if err := filetool.AppendString(*file, *data+"\n"); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(config.ExitIOError)
		}
		fmt.Printf("✓ 已追加 1 行到 %s\n", *file)

	case "list":
		root := *file
		if root == "" {
			root = "."
		}
		entries, err := filetool.WalkFiltered(root,
			[]string{".git", "node_modules", "dist"})
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(config.ExitIOError)
		}
		for _, e := range entries {
			if e.IsDir {
				fmt.Printf("📁 %s/\n", e.Path)
			} else {
				fmt.Printf("📄 %s (%d B)\n", e.Path, e.Size)
			}
		}

	case "grep":
		var (
			results []filetool.GrepResult
			err     error
		)

		if *regex {
			results, err = filetool.GrepRegex(*file, *pattern)
		} else {
			results, err = filetool.GrepContains(*file, *pattern)
		}

		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(config.ExitIOError)
		}

		for _, r := range results {
			fmt.Printf("%d: %s\n", r.LineNum, r.Line)
		}
		fmt.Printf("\n共匹配 %d 行\n", len(results))

	case "embed":
		names, err := filetool.ListEmbedded()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(config.ExitIOError)
		}
		fmt.Println("嵌入的资源文件:")
		for _, n := range names {
			fmt.Println("  ", n)
		}

	case "help":
		fmt.Print(filetool.HelpText)

	default:
		fmt.Fprintln(os.Stderr,
			"用法: file -action read|write|append|list|grep|embed|help [-file path] [-pattern str] [-data str] [-regex]")
		os.Exit(config.ExitUsageError)
	}
}
