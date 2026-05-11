// demo/filetool_demo/main.go
// 第2章所有文件 IO 知识点的统一演示程序。
// 运行方式：go run ./demo/filetool_demo
package main

import (
	"bufio"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func main() {
	tmpDir := setupSandbox()
	defer os.RemoveAll(tmpDir)

	fmt.Println("沙盒目录", tmpDir)

	demoReadWrite(tmpDir)
	demoLineFilter(tmpDir)
	demoFilePaths(tmpDir)
	demoDirectoryWalk(tmpDir)
	demoTempFile()
}

// setupSandbox 创建一个独立的临时目录用作演示沙盒，避免污染项目目录。
func setupSandbox() string {
	dir, err := os.MkdirTemp("", "filetool-demo-*")
	if err != nil {
		panic(err)
	}
	return dir
}

// 2.1+2.2 读写演示
func demoReadWrite(dir string) {
	fmt.Println("\n=== 2.1+2.2 文件读写 ===")
	path := filepath.Join(dir, "hello.txt")

	// 覆盖写
	os.WriteFile(path, []byte("Hello, DevOps!\n"), 0644)
	fmt.Println("✓ 写入 hello.txt")

	// 整读
	data, _ := os.ReadFile(path)
	fmt.Printf("整读结果: %q\n", string(data))

	// 追加写
	f, _ := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0644)
	f.WriteString("Second line\n")
	f.WriteString(fmt.Sprintf("Time: %s\n", time.Now().Format("15:04:05")))
	f.Close()

	// 流式读
	fmt.Println("流式读取（按行）:")
	f2, _ := os.Open(path)
	defer f2.Close()
	sc := bufio.NewScanner(f2)
	for sc.Scan() {
		fmt.Printf("  | %s\n", sc.Text())
	}
}

// 2.3 行过滤演示
func demoLineFilter(dir string) {
	fmt.Println("\n=== 2.3 行过滤（grep）===")
	path := filepath.Join(dir, "log.txt")
	os.WriteFile(path, []byte(strings.Join([]string{
		"2025-05-08 10:00:01 INFO  服务启动",
		"2025-05-08 10:00:05 DEBUG 处理请求",
		"2025-05-08 10:00:10 ERROR 数据库连接失败",
		"2025-05-08 10:00:15 INFO  重试中",
		"2025-05-08 10:00:20 ERROR 超时",
	}, "\n")), 0644)

	f, _ := os.Open(path)
	defer f.Close()
	sc := bufio.NewScanner(f)
	lineNum := 0
	for sc.Scan() {
		lineNum++
		if strings.Contains(sc.Text(), "ERROR") {
			fmt.Printf("⚠️  行 %d: %s\n", lineNum, sc.Text())
		}
	}
}

// 2.4 路径工具演示
func demoFilePaths(dir string) {
	fmt.Println("\n=== 2.4 路径工具 ===")
	p := filepath.Join(dir, "subdir", "report.tar.gz")

	fmt.Printf("Join     : %s\n", p)
	fmt.Printf("Base     : %s\n", filepath.Base(p))
	fmt.Printf("Dir      : %s\n", filepath.Dir(p))
	fmt.Printf("Ext      : %s\n", filepath.Ext(p))
	fmt.Printf("IsAbs    : %v\n", filepath.IsAbs(p))

	abs, _ := filepath.Abs(".")
	fmt.Printf("当前绝对 : %s\n", abs)

	// Glob 模式匹配
	matches, _ := filepath.Glob(filepath.Join(dir, "*.txt"))
	fmt.Printf("Glob *.txt: %d 个 -> %v\n", len(matches), matches)
}

// 2.5 目录遍历演示
func demoDirectoryWalk(dir string) {
	fmt.Println("\n=== 2.5 目录遍历 ===")

	// 先建几层目录
	os.MkdirAll(filepath.Join(dir, "a", "b", "c"), 0755)
	os.WriteFile(filepath.Join(dir, "a", "x.txt"), []byte("x"), 0644)
	os.WriteFile(filepath.Join(dir, "a", "b", "y.txt"), []byte("yy"), 0644)
	os.WriteFile(filepath.Join(dir, "a", "b", "c", "z.txt"), []byte("zzz"), 0644)

	// Walk 遍历
	filepath.Walk(filepath.Join(dir, "a"), func(path string, info fs.FileInfo, err error) error {
		depth := strings.Count(strings.TrimPrefix(path, dir), string(os.PathSeparator))
		indent := strings.Repeat("  ", depth)

		if info.IsDir() {
			fmt.Printf("%s📁 %s/\n", indent, info.Name())
		} else {
			fmt.Printf("%s📄 %s (%d B)\n", indent, info.Name(), info.Size())
		}
		return nil
	})
}

// 2.6 临时文件演示
func demoTempFile() {
	fmt.Println("\n=== 2.6 临时文件 ===")

	tmp, _ := os.CreateTemp("", "demo-*.txt")
	defer os.Remove(tmp.Name())
	tmp.WriteString(fmt.Sprintf("时间: %s\n", time.Now()))
	tmp.Close()

	fmt.Printf("临时文件: %s\n", tmp.Name())
	data, _ := os.ReadFile(tmp.Name())
	fmt.Printf("内容: %q\n", string(data))

	tmpDir, _ := os.MkdirTemp("", "demo-dir-*")
	defer os.RemoveAll(tmpDir)
	fmt.Printf("临时目录: %s\n", tmpDir)
}
