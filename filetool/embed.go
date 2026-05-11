// 概念：`//go:embed` 指令把外部文件编译进二进制，发布时无需带 resource 目录。

// filetool/embed.go
// //go:embed 演示：把 resources/ 整个目录嵌入二进制。
// 三种嵌入类型：
//
//	//go:embed file.txt        → string  或 []byte  (单文件)
//	//go:embed dir/*           → embed.FS           (多文件 / 整个目录)
//	//go:embed dir1 dir2 *.md  → embed.FS           (混合)
package filetool

import (
	"embed"
	"fmt"
)

// resourcesFS 是嵌入的虚拟文件系统。
//
// 注意：//go:embed 指令必须紧贴在 var 声明上面，中间不能有空行。
//
//go:embed resources/*
var resourcesFS embed.FS

// HelpText 直接以 string 形式嵌入单个文件。
//
//go:embed resources/help.txt
var HelpText string

// ListEmbedded 列出嵌入的所有资源文件名。
func ListEmbedded() ([]string, error) {
	entries, err := resourcesFS.ReadDir("resources")
	if err != nil {
		return nil, fmt.Errorf("读取嵌入目录失败: %w", err)
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		names = append(names, e.Name())
	}
	return names, nil
}

// ReadEmbedded 读取嵌入的任意文件。
// 例：ReadEmbedded("resources/help.txt") → ([]byte, nil)
func ReadEmbedded(path string) ([]byte, error) {
	return resourcesFS.ReadFile(path)
}
