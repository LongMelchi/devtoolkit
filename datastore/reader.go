// 设计说明：把"读取并解析 jsonl"的能力独立出来，便于 `data tail` 子命令和后续的统计聚合复用。

// datastore/reader.go
//
// 提供"按行流式读取并解析 jsonl"的工具函数。
//
// 设计：
//   - 不一次性把整个文件读进内存，而是用 bufio.Scanner 逐行读
//   - 单行 JSON 解析失败时不中断整体流程，而是返回错误供调用方决定如何处理
package datastore

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// ReadAll 读取一个 .jsonl 文件并返回其中所有 Event。
//
// 大文件慎用：会一次性把所有 Event 装入切片。
// 大文件请用 ReadEach 流式回调版本。
func ReadAll(path string) ([]*Event, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("ReadAll: 打开 %s 失败: %w", path, err)
	}
	defer f.Close()

	var events []*Event
	scanner := bufio.NewScanner(f)
	// jsonl 单行可能包含较长 payload，扩大 Scanner 缓冲到 1MB
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)

	lineNo := 0
	for scanner.Scan() {
		lineNo++
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		var e Event
		if err := json.Unmarshal(line, &e); err != nil {
			return events, fmt.Errorf("ReadAll: 第 %d 行 JSON 解析失败: %w", lineNo, err)
		}
		events = append(events, &e)
	}
	if err := scanner.Err(); err != nil && err != io.EOF {
		return events, fmt.Errorf("ReadAll: 扫描失败: %w", err)
	}
	return events, nil
}

// ReadEach 流式读取，每解析出一个 Event 就回调一次 fn。
//
// 当 fn 返回 false 时停止扫描（适合 `tail -n N` 之类的提前终止）。
// 推荐处理大文件（>100MB）时使用，避免大切片占用内存。
func ReadEach(path string, fn func(*Event) bool) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		var e Event
		if err := json.Unmarshal(line, &e); err != nil {
			continue // 跳过损坏行，保证流式读取的鲁棒性
		}
		if !fn(&e) {
			return nil
		}
	}
	return scanner.Err()
}

// ListFiles 列出 dataDir 下属于某个 module 的所有 .jsonl 文件，按时间倒序。
//
// 当 module 为 "" 时列出所有模块的文件。
func ListFiles(dataDir, module string) ([]string, error) {
	entries, err := os.ReadDir(dataDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // 目录不存在视为空列表，符合 ls 习惯
		}
		return nil, err
	}

	var files []string
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".jsonl") {
			continue
		}
		if module != "" && !strings.HasPrefix(name, module+"-") {
			continue
		}
		files = append(files, filepath.Join(dataDir, name))
	}
	sort.Sort(sort.Reverse(sort.StringSlice(files))) // 文件名按日期排序，倒序 = 新的在前

	return files, nil
}
