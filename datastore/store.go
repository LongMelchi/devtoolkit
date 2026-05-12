// datastore/store.go
//
// Store 是 datastore 的"写入端"，负责把 Event 以 JSON Lines 形式追加到
// 指定模块的 .jsonl 文件中。
//
// 文件命名规则：<DataDir>/<module>-YYYY-MM-DD.jsonl
//
//	例如：./data/file-2026-05-08.jsonl
//
// 并发安全：单进程内通过互斥锁保护 Encoder；跨进程依赖 OS 的 O_APPEND 原子性
// （Linux/macOS：< PIPE_BUF 字节的写入是原子的；Windows 在多数情况下也是原子）。
package datastore

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Store 表示一个打开的 datastore 写入会话。
type Store struct {
	dir     string        // 数据根目录（来自 config.DataDir）
	module  string        // 当前 Store 绑定的模块名
	file    *os.File      // 当前打开的 .jsonl 文件
	encoder *json.Encoder // 复用的编码器，避免每次写都新建对象
	mu      sync.Mutex    // 保护 Write，避免多 goroutine 交错写入
}

// Open 打开（或创建）某模块今天的 .jsonl 文件，返回 Store。
//
// 调用方必须在使用完毕后调用 Close。
//
// 参数：
//
//	dir    数据根目录（建议传 cfg.DataDir，例如 "./data"）
//	module 模块名，用于组成文件名（例如 "file" / "net" / "healthcheck"）
func Open(dir, module string) (*Store, error) {
	if dir == "" || module == "" {
		return nil, fmt.Errorf("datastore.Open: dir 和 module 都不能为空")
	}

	// 1. 确保目录存在（递归创建）
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("datastore.Open: 创建目录 %s 失败: %w", dir, err)
	}

	// 2. 文件名：<module>-YYYY-MM-DD.jsonl
	filename := fmt.Sprintf("%s-%s.jsonl", module, time.Now().Format("2006-01-02"))
	path := filepath.Join(dir, filename)

	// 3. 以 O_APPEND | O_CREATE | O_WRONLY 模式打开 ★★★ 本章核心
	//
	//    O_APPEND: 每次 Write 自动定位到末尾（多进程下也安全）
	//    O_CREATE: 文件不存在则创建
	//    O_WRONLY: 只写不读（提升一点性能，且明确意图）
	//    0o644:    权限位 rw-r--r--（owner 读写，其他人只读）
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, fmt.Errorf("datastore.Open: 打开 %s 失败: %w", path, err)
	}

	return &Store{
		dir:     dir,
		module:  module,
		file:    f,
		encoder: json.NewEncoder(f), // Encoder 自动在每个 Encode 后输出换行符 \n
	}, nil

}

// Append 把一个 Event 追加到当前文件，自动写入换行符。
//
// 这是 datastore 的核心 API：上层模块只需调 NewEvent + Append。
func (s *Store) Append(e *Event) error {
	if e == nil {
		return fmt.Errorf("datastore.Append: event 不能为 nil")
	}

	if e.Module == "" {
		e.Module = s.module // 调用方未填则用 Store 绑定的 module
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// json.Encoder.Encode 会自动在末尾追加 '\n'，正合 JSON Lines 格式
	if err := s.encoder.Encode(e); err != nil {
		return fmt.Errorf("datastore.Append: 编码失败: %w", err)
	}

	return nil
}

// Close 关闭文件句柄。defer 调用即可。
func (s *Store) Close() error {
	if s.file == nil {
		return nil
	}
	return s.file.Close()
}

// Path 返回当前 Store 写入的文件绝对/相对路径，便于日志展示。
func (s *Store) Path() string {
	if s.file == nil {
		return ""
	}
	return s.file.Name()
}
