// procman/spawner.go
//
// 派生进程：cmd.Start() 让主进程立即返回拿到 PID，
// 后续可主动终止或调用 Wait 同步等结束。
//
// 这是 procmon 模块（第6.3章）"进程崩溃自动重启"的基础。
package procman

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

// SpawnedProcess 表示一个已派生的子进程。
type SpawnedProcess struct {
	cmd     *exec.Cmd
	PID     int
	Command string
}

// Spawn 启动子进程，立即返回，不等其结束。
//
// extraEnv 会附加在父进程环境之上（os.Environ() 自带 PATH 等基础变量）。
// 可用于传递 DEVTOOLKIT_HOST=... 之类。
//
// stdoutSink/stderrSink 可以是 os.Stdout/os.Stderr，也可以是文件。
// 传 nil 表示 /dev/null（丢弃）。
func Spawn(extraEnv []string, stdoutSink, stderrSink io.Writer, name string, args ...string) (*SpawnedProcess, error) {
	cmd := exec.Command(name, args...)

	// 环境变量：先继承父进程，再追加传入的
	cmd.Env = append(os.Environ(), extraEnv...)

	// stdin 接管父进程，便于交互式子进程；不需要时可设为 nil
	cmd.Stdin = os.Stdin
	cmd.Stdout = stdoutSink
	cmd.Stderr = stderrSink

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("启动 %q 失败: %w", name, err)
	}
	return &SpawnedProcess{
		cmd:     cmd,
		PID:     cmd.Process.Pid,
		Command: name + " " + strings.Join(args, " "),
	}, nil
}

// Wait 阻塞等待子进程结束，返回退出码。
func (p *SpawnedProcess) Wait() (exitCode int, err error) {
	err = p.cmd.Wait()
	if p.cmd.ProcessState != nil {
		exitCode = p.cmd.ProcessState.ExitCode()
	}
	return exitCode, err
}

// Kill 立即终止子进程（SIGKILL）。
//
// 注意：SIGKILL 不可被捕获，子进程没有清理机会。
// 优雅关闭应该用 Signal(syscall.SIGTERM) + 等待。
func (p *SpawnedProcess) Kill() error {
	if p.cmd.Process == nil {
		return nil
	}
	return p.cmd.Process.Kill()
}

// IsRunning 检查进程是否仍在运行（启发式）。
//
// 在类 Unix 系统上，Process.Signal(syscall.Signal(0)) 是检测进程存活的标准技巧：
// 不发送任何信号，但若进程不存在会返回错误。
func (p *SpawnedProcess) IsRunning() bool {
	if p.cmd.ProcessState != nil && p.cmd.ProcessState.Exited() {
		return false
	}
	if p.cmd.Process == nil {
		return false
	}

	// signal 0 是探测信号，跨平台行为略有差异
	return p.cmd.Process.Signal(os.Signal(nil)) == nil ||
		p.cmd.ProcessState == nil
}
