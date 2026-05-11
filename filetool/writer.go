// filetool/writer.go
// 文件写入工具：覆盖写、追加写、原子写。
package filetool

import (
	"fmt"
	"os"
	"path/filepath"
)

// WriteFile 覆盖写文件（如不存在则创建）。
// 权限说明：
//
//	0644 = rw-r--r-- (Unix)：所有者读写，其他人只读
//	0600 = rw------- ：仅所有者读写（用于敏感文件如配置/密钥）
//	0755 = rwxr-xr-x ：可执行文件
//
// Windows 下权限位被忽略，仅保留可读/可写。
func WriteFile(path string, data []byte) error {
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("写入 %q 失败: %w", path, err)
	}
	return nil
}

// AppendString 追加字符串到文件末尾（不存在则创建）。
// OpenFile 标志可位或组合，常见组合：
//
//	O_APPEND|O_CREATE|O_WRONLY  追加写（不清空已有内容）
//	O_TRUNC |O_CREATE|O_WRONLY  覆盖写（先清空）
//	O_CREATE|O_RDWR             读写打开（不清空）
func AppendString(path, line string) error {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("打开 %q 追加写失败: %w", path, err)
	}
	defer f.Close()

	if _, err := f.WriteString(line); err != nil {
		return fmt.Errorf("追加内容到 %q 失败: %w", path, err)
	}
	return nil
}

// WriteAtomic 原子写：先写到临时文件，再 rename 到目标路径。
// 用途：避免"写到一半被 kill"导致目标文件半截损坏。
// 原理：rename 在同分区是原子操作，要么完整切换，要么不变。
func WriteAtomic(path string, data []byte) error {
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".tmp-*")
	if err != nil {
		return fmt.Errorf("创建临时文件失败: %w", err)
	}
	tmpPath := tmp.Name()

	// 写入临时文件
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		os.Remove(tmpPath) // 清理失败的临时文件
		return fmt.Errorf("写临时文件失败: %w", err)
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("关闭临时文件失败: %w", err)
	}

	// 原子替换：等价于 mv tmp path
	if err := os.Rename(tmpPath, path); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("重命名 %q -> %q 失败: %w", tmpPath, path, err)
	}
	return nil
}
