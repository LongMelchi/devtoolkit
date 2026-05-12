// 设计说明：长期运行可能让单个 jsonl 文件变得很大。
// 本文件提供一个最小化的"按日切分"辅助器，演示文件命名规范与日期判断。
// datastore/rotator.go
//
// 简易按日切分：判断当前 Store 的文件是否属于"今天"，否则提示需重新 Open。
//
// 完整的日志切分（按大小、按数量、压缩归档）超出本章范围，
// 此处仅演示判断逻辑，让学习者理解"切分时机"这个概念。
package datastore

import (
	"path/filepath"
	"strings"
	"time"
)

// IsToday 判断给定 jsonl 文件名是否包含今天的日期。
//
// 文件命名约定：<module>-YYYY-MM-DD.jsonl
// 实现思路：取文件名末尾的日期段，与 time.Now() 比较。
func IsToday(path string) bool {
	base := filepath.Base(path)
	base = strings.TrimSuffix(base, ".jsonl")
	// 例如 base = "file-2026-05-08"
	parts := strings.Split(base, "-")
	if len(parts) < 4 {
		return false
	}
	dateStr := strings.Join(parts[len(parts)-3:], "-") // 取最后 3 段
	today := time.Now().Format("2006-01-02")
	return dateStr == today
}

// ShouldRotate 给定一个 Store，判断是否需要重新 Open（跨天了）。
//
// 调用方典型用法：
//
//	if datastore.ShouldRotate(store) {
//	    store.Close()
//	    store, _ = datastore.Open(cfg.DataDir, "file")
//	}
func ShouldRotate(s *Store) bool {
	if s == nil || s.file == nil {
		return false
	}
	return !IsToday(s.file.Name())
}
