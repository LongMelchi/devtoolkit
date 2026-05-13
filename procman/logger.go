//    logger.Println("hello")
//             │
//             ▼ 调用 io.Writer.Write([]byte)
//    ┌──────────────────────┐
//    │  io.MultiWriter      │   ← 把一次 Write 复制到多个目标
//    └──┬──────────────┬────┘
//       │              │
//       ▼              ▼
//    os.Stdout      os.File (app.log)
//    "hello\n"      "hello\n"

//    ✓ 一次写入，多处可见
//    ✓ 可以方便地组合：MultiWriter(stdout, file1, file2, network)

// procman/logger.go
//
// 日志工具：构造同时输出到控制台和文件的 logger。
//
// 设计要点：
//   - 不污染全局 log 包，每个调用方拥有独立 *log.Logger
//   - Close() 必须调用，否则文件句柄泄漏
package procman

import (
	"fmt"
	"io"
	"log"
	"os"
)

// Logger 是带文件句柄的 logger 包装。
type Logger struct {
	*log.Logger          // 结构体内嵌（Embedding）
	file        *os.File // 持有文件句柄以便 Close
}

// NewLogger 创建一个 logger，同时输出到 stdout 和指定文件。
//
// 参数：
//
//	prefix  - 日志前缀，如 "[devtool] "
//	logFile - 日志文件路径，空字符串表示只输出到 stdout
//
// 文件以 O_APPEND 方式打开，多次启动会累加写入而非覆盖。
func NewLogger(prefix, logFile string) (*Logger, error) {
	var writers []io.Writer = []io.Writer{os.Stdout}
	var file *os.File

	if logFile != "" {
		f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return nil, fmt.Errorf("打开日志文件 %q 失败: %w", logFile, err)
		}
		file = f
		writers = append(writers, f)
	}

	multi := io.MultiWriter(writers...)

	// log.Lshortfile 输出文件名+行号，便于定位调用处
	// log.LstdFlags 输出 "yyyy/mm/dd hh:mm:ss"
	lg := log.New(multi, prefix, log.LstdFlags|log.Lshortfile)

	return &Logger{Logger: lg, file: file}, nil
}

// Close 关闭日志文件（若有）。
func (l *Logger) Close() error {
	if l.file != nil {
		return l.file.Close()
	}
	return nil
}

// SetDebug 切换到带微秒精度的 debug 日志格式。
//
// 用法：在程序启动时根据 DEVTOOLKIT_LOG=debug 调用本方法。
func (l *Logger) SetDebug(enable bool) {
	if enable {
		l.SetFlags(log.LstdFlags | log.Lmicroseconds | log.Lshortfile)
		l.SetPrefix("[DEBUG] ")
	} else {
		l.SetFlags(log.LstdFlags | log.Lshortfile)
		l.SetPrefix("[INFO] ")
	}
}
