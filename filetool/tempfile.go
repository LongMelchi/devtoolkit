// filetool/tempfile.go
// 临时文件管理：用于"写到一半要做处理，最后再决定是否保留"的场景。
package filetool

import (
	"fmt"
	"os"
)

// TempFile 是临时文件的简单封装，把 cleanup 函数和路径捆绑。
// 设计意图：调用方只需要拿到 path 和 cleanup，不必关心底层 *os.File。
type TempFile struct {
	Path    string
	Cleanup func() // 调用方必须 defer 它（或在合适的时机调用）
}

// CreateTemp 创建一个唯一命名的临时文件。
// dir       - 临时目录；空字符串表示系统默认（Linux: /tmp, Windows: %TEMP%）
// pattern   - 文件名模板，* 会被随机字符串替换
//
//	例如 "devtool-*.txt" 可能生成 "devtool-1234567890.txt"
//
// 返回的 TempFile.Cleanup 包含了关闭和删除两步。
func CreateTemp(dir, pattern string) (*TempFile, error) {
	f, err := os.CreateTemp(dir, pattern)
	if err != nil {
		return nil, fmt.Errorf("创建临时文件失败: %w", err)
	}
	path := f.Name()
	_ = f.Close() // 立即关闭，调用方只通过路径操作

	return &TempFile{
		Path: path,
		Cleanup: func() {
			os.Remove(path)
		},
	}, nil
}

// CreateTempDir 创建一个唯一命名的临时目录。
// 类似 CreateTemp，但 Cleanup 用 RemoveAll 递归删除。
func CreateTempDir(dir, pattern string) (*TempFile, error) {
	path, err := os.MkdirTemp(dir, pattern)
	if err != nil {
		return nil, fmt.Errorf("创建临时目录失败: %w", err)
	}
	return &TempFile{
		Path: path,
		Cleanup: func() {
			os.RemoveAll(path) // 递归删除
		},
	}, nil
}
