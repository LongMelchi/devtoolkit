// filetool/pathutil.go
// 路径工具：跨平台拼接、解析、匹配。
// 核心：path/filepath 包自动处理 Windows (\) vs Unix (/) 分隔符。
// 不要用字符串拼接 path1 + "/" + path2！
package filetool

import (
	"path/filepath"
	"strings"
)

// PathInfo 包含路径的常用解析结果。
// 设计：把多次调用 filepath 各方法的结果合并到一个结构体，
// 调用方一次拿全所有信息。
type PathInfo struct {
	Original  string // 原始路径
	Absolute  string // 绝对路径
	Dir       string // 目录部分
	Base      string // 文件名（含扩展名）
	NameOnly  string // 文件名（不含扩展名）
	Extension string // 扩展名（含 .）
	IsAbs     bool   // 是否为绝对路径
}

// Inspect 解析路径，返回 PathInfo。
// 示例：Inspect("./logs/app.log") 在 Linux 下返回
//
//	Original: "./logs/app.log"
//	Absolute: "/home/user/proj/logs/app.log"
//	Dir:      "logs"
//	Base:     "app.log"
//	NameOnly: "app"
//	Extension:".log"
func Inspect(path string) PathInfo {
	abs, _ := filepath.Abs(path) // 失败也无所谓，返回空字符串
	base := filepath.Base(path)
	ext := filepath.Ext(path)
	return PathInfo{
		Original:  path,
		Absolute:  abs,
		Dir:       filepath.Dir(path),
		Base:      base,
		NameOnly:  strings.TrimSuffix(base, ext),
		Extension: ext,
		IsAbs:     filepath.IsAbs(path),
	}
}

// Join 跨平台路径拼接。
// 简单转发到 filepath.Join，导出主要是为了让上层只 import 一个包。
func Join(elem ...string) string {
	return filepath.Join(elem...)
}

// Glob 模式匹配：找到所有匹配的文件。
//
// 例：Glob("*.go") 返回当前目录所有 .go 文件
//
//	Glob("logs/*.log") 返回 logs/ 下所有 .log 文件
//
// 注意：不支持递归 ** 通配符，递归匹配请用 filepath.Walk + 自己过滤。
func Glob(pattern string) ([]string, error) {
	return filepath.Glob(pattern)
}
