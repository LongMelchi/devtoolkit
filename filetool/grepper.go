// filetool/grepper.go
// 行过滤：实现简单版的 grep 功能。
// 核心组件：bufio.Scanner（默认按行切分，缓冲读取，性能优秀）。
package filetool

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
)

// GrepResult 表示一行匹配结果。
type GrepResult struct {
	LineNum int    // 行号（从 1 开始）
	Line    string // 行内容（已去除末尾换行）
}

// GrepContains 在 path 中查找包含 keyword 的行。
// 这是最简单的子串匹配版本。
func GrepContains(path, keyword string) ([]GrepResult, error) {
	return grepWith(path, func(line string) bool {
		return strings.Contains(line, keyword)
	})
}

// GrepRegex 用正则表达式查找。
// 例：GrepRegex("app.log", `ERROR|FATAL`) 匹配错误日志。
func GrepRegex(path, pattern string) ([]GrepResult, error) {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("正则编译失败 %q: %w", pattern, err)
	}
	return grepWith(path, re.MatchString)
}

// grepWith 是私有的通用扫描器：对每行调用 match() 决定是否保留。
// 这是"高阶函数"模式：把"如何匹配"作为参数传入，复用扫描骨架。
// 类似项目2 中的 sortBy(items, less func(a, b) int)。
func grepWith(path string, match func(line string) bool) ([]GrepResult, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("打开 %q 失败: %w", path, err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	// 默认 buffer 64KB，超长单行会报 "token too long"。
	// 处理日志类文件时建议扩大：
	// scanner.Buffer(make([]byte, 1024*1024), 10*1024*1024) // 1MB 起，最大 10MB

	var results []GrepResult
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := scanner.Text() // Text() 自动去除末尾的 \n 和 \r
		if match(line) {
			results = append(results, GrepResult{LineNum: lineNum, Line: line})
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("扫描 %q 时出错: %w", path, err)
	}
	return results, nil
}
