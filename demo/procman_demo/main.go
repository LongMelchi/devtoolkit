// demo/procman_demo/main.go
//
// 第4章进程管理与日志演示。
// 运行方式：go run ./demo/procman_demo
//
// 兼容性：会根据 GOOS 选择 Windows/Unix 适用的命令。
package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

func main() {
	demoExec()
	demoExecOutput()
	demoExecStdin()
	demoSpawn()
	demoLogger()
}

// 跨平台地选择简单可用的命令
func pickListCmd() (string, []string) {
	if runtime.GOOS == "windows" {
		return "cmd", []string{"/c", "dir"}
	}
	return "ls", []string{"-la"}
}

// 4.1 cmd.Run 同步执行
func demoExec() {
	fmt.Println("\n=== 4.1 cmd.Run 同步执行 ===")
	name, args := pickListCmd()
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Println("执行失败:", err)
		return
	}
	fmt.Printf("[执行完毕，退出码=%d]\n", cmd.ProcessState.ExitCode())
}

// 4.1b cmd.Output 捕获输出
func demoExecOutput() {
	fmt.Println("\n=== 4.1b cmd.Output 捕获输出 ===")
	out, err := exec.Command("go", "version").Output()
	if err != nil {
		fmt.Println("失败:", err)
		return
	}
	fmt.Println("Go 版本:", strings.TrimSpace(string(out)))
}

// 4.1c stdin 喂入数据
func demoExecStdin() {
	fmt.Println("\n=== 4.1c stdin 喂数据 ===")
	if runtime.GOOS == "windows" {
		fmt.Println("(跳过：Windows 没有现成的字数统计命令)")
		return
	}
	cmd := exec.Command("wc", "-w")
	cmd.Stdin = strings.NewReader("hello world from go demo")
	out, _ := cmd.Output()
	fmt.Printf("wc 统计字数: %s", out)
}

// 4.2 cmd.Start + cmd.Wait 派生
func demoSpawn() {
	fmt.Println("\n=== 4.2 cmd.Start + cmd.Wait 派生进程 ===")
	var name string
	var args []string

	if runtime.GOOS == "windows" {
		name, args = "cmd", []string{"/c", "echo Hello from child & timeout /t 1"}
	} else {
		name, args = "sh", []string{"-c", "echo Hello from child; sleep 1"}
	}

	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	if err := cmd.Start(); err != nil {
		fmt.Println("启动失败:", err)
		return
	}
	fmt.Printf("👶 子进程 PID=%d，主进程可继续做别的事...\n", cmd.Process.Pid)
	time.Sleep(200 * time.Millisecond)
	fmt.Println("📌 主进程做的事：等待...")
	if err := cmd.Wait(); err != nil {
		fmt.Println("子进程异常退出:", err)
	} else {
		fmt.Println("✅ 子进程正常退出")
	}
}

// 4.3 io.MultiWriter 多路日志
func demoLogger() {
	fmt.Println("\n=== 4.3 io.MultiWriter 多路日志 ===")
	tmp, _ := os.CreateTemp("", "demo-log-*.log")
	defer os.Remove(tmp.Name())

	multi := io.MultiWriter(os.Stdout, tmp)
	lg := log.New(multi, "[demo] ", log.LstdFlags)
	lg.Println("第 1 条日志")
	lg.Printf("格式化日志: port=%d", 9090)

	tmp.Close()

	// 验证文件中也写入了
	data, _ := os.ReadFile(tmp.Name())
	fmt.Println("\n--- 文件中的内容 ---")
	fmt.Print(string(data))
	fmt.Println("--------------------")

	// 兼顾验证：用一个新 buffer 写日志
	var buf bytes.Buffer
	lg2 := log.New(&buf, "[buf] ", log.LstdFlags)
	lg2.Println("写入到 bytes.Buffer 的日志")
	fmt.Print("Buffer 内容: ", buf.String())
}
