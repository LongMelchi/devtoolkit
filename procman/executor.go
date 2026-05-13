// procman/executor.go
//
// 执行外部命令的工具：同步执行、捕获输出、传 stdin。
//
// 核心组件：os/exec 包。这是 Go 操作外部进程的标准方式。
package procman

import (
	"bytes"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"time"
)

// ExecResult 是一次命令执行的完整结果。
type ExecResult struct {
	Command  string        // 实际执行的命令字符串（含参数）
	Stdout   string        // 标准输出内容
	Stderr   string        // 标准错误内容
	ExitCode int           // 退出码（成功为 0）
	Duration time.Duration // 执行耗时
	Err      error         // 执行错误（命令找不到等）
}

// IsSuccess 判断是否成功执行（退出码 0 且无错误）。
func (r *ExecResult) IsSuccess() bool {
	return r.Err == nil && r.ExitCode == 0
}

// Run 同步执行命令并捕获输出。
//
// name 是命令名（如 "go" 或 "ls"），args 是参数列表。
//
// 设计要点：
//   - 用独立的 stdout/stderr buffer 而非共享，便于分别处理
//   - 捕获退出码：cmd.ProcessState.ExitCode() 或从 ExitError 提取
//   - 计时用 time.Now() 简单准确
func Run(name string, args ...string) *ExecResult {
	start := time.Now()
	cmd := exec.Command(name, args...)

	// 重定向输出到缓冲区
	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	err := cmd.Run()

	result := &ExecResult{
		Command:  name + " " + strings.Join(args, " "),
		Stdout:   outBuf.String(),
		Stderr:   errBuf.String(),
		Duration: time.Since(start),
		Err:      err,
	}

	// 提取退出码
	if cmd.ProcessState != nil {
		result.ExitCode = cmd.ProcessState.ExitCode()
	} else if err != nil {
		result.ExitCode = -1 // 命令完全没启动起来
	}
	return result
}

// RunWithStdin 执行命令并通过 stdin 喂入数据。
//
// 例：RunWithStdin("hello world\n", "wc", "-w") → 输出 "2"
func RunWithStdin(stdin string, name string, args ...string) *ExecResult {
	start := time.Now()
	cmd := exec.Command(name, args...)
	cmd.Stdin = strings.NewReader(stdin)

	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	err := cmd.Run()
	result := &ExecResult{
		Command:  name + " " + strings.Join(args, " "),
		Stdout:   outBuf.String(),
		Stderr:   errBuf.String(),
		Duration: time.Since(start),
		Err:      err,
	}

	if cmd.ProcessState != nil {
		result.ExitCode = cmd.ProcessState.ExitCode()
	}
	return result
}

// RunStreaming 执行命令并把 stdout 实时写到 sink（如 os.Stdout）。
//
// 适用场景：长时间运行的命令（如 docker logs -f），希望实时看到输出而非等结束。
func RunStreaming(sink io.Writer, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = sink
	cmd.Stderr = sink

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("命令 %q 执行失败: %w", name, err)
	}
	return nil
}
