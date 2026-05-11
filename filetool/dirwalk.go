// filetool/dirwalk.go
// 目录遍历：递归扫描整个目录树。
// 选用 filepath.Walk 而非更新的 filepath.WalkDir 是因为前者
// 直接传 FileInfo（包含大小等），WalkDir 性能更好但需要再 Stat。
// 本项目教学场景下可读性优先。
package filetool

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

// FileEntry 描述一个被遍历到的文件或目录。
type FileEntry struct {
	Path  string // 相对/绝对路径（取决于调用方传入的 root）
	IsDir bool
	Size  int64       // 字节数（目录的 Size 没有意义）
	Mode  os.FileMode // 权限位
}

// EnsureDir 确保目录存在；不存在则递归创建（类似 mkdir -p）。
// MkdirAll 已经处理了"目录已存在"的情况，不会报错。
func EnsureDir(path string) error {
	if err := os.MkdirAll(path, 0755); err != nil {
		return fmt.Errorf("创建目录 %q 失败: %w", path, err)
	}
	return nil
}

// WalkAll 遍历 root 下所有文件和目录，返回扁平列表。
// 用 filepath.Walk 而非 os.ReadDir，因为它已经帮我们做了递归。
func WalkAll(root string) ([]FileEntry, error) {
	var entries []FileEntry

	walkFn := func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			// info 可能为 nil，比如权限不足、目录不存在
			return fmt.Errorf("访问 %q 出错: %w", path, err)
		}
		entries = append(entries, FileEntry{
			Path:  path,
			IsDir: info.IsDir(),
			Size:  info.Size(),
			Mode:  info.Mode(),
		})
		return nil
	}

	if err := filepath.Walk(root, walkFn); err != nil {
		return nil, err
	}
	return entries, nil
}

// WalkFiltered 遍历但跳过不感兴趣的目录（如 .git, node_modules）。
// 通过返回 filepath.SkipDir 实现"跳过该目录但继续遍历其他"。
func WalkFiltered(root string, skipDirs []string) ([]FileEntry, error) {
	skipSet := make(map[string]struct{}, len(skipDirs))
	for _, d := range skipDirs {
		skipSet[d] = struct{}{}
	}

	var entries []FileEntry
	walkFn := func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			if _, skip := skipSet[info.Name()]; skip {
				return filepath.SkipDir // 关键魔法：返回 SkipDir 跳过整个子树
			}
		}
		entries = append(entries, FileEntry{
			Path: path, IsDir: info.IsDir(), Size: info.Size(), Mode: info.Mode(),
		})
		return nil
	}
	return entries, filepath.Walk(root, walkFn)
}
