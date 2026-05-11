// filetool/reader.go
// 文件读取工具：提供"整读"和"流式读取"两种模式。
// 设计原则：
//   - 函数签名都返回 (result, error)，由调用方决定如何处理错误
//   - 不直接调用 os.Exit，避免污染上层代码的退出策略
package filetool

import (
	"fmt"
	"io"
	"os"
)

// ReadAll 一次性读取整个文件并返回字节内容。
// 适用场景：配置文件、小型文本文件（< 10MB）。
// 错误处理：把 os.ReadFile 的底层错误用 fmt.Errorf 包装一层，
// 加上文件名上下文，便于调用方定位问题。
func ReadAll(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		// 格式化消息或包装错误
		return nil, fmt.Errorf("读取文件 %q 失败：%w", path, err)
	}
	return data, nil
}

// ReadAllString 是 ReadAll 的字符串版本，便于直接处理文本。
func ReadAllString(path string) (string, error) {
	data, err := ReadAll(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// ReadStream 流式读取大文件：每次读 chunkSize 字节，回调 onChunk 处理。
// 参数：
//
//	path      - 文件路径
//	chunkSize - 每次读取的字节数（建议 4096 或 8192，对齐磁盘块）
//	onChunk   - 处理回调；返回 error 中断读取
//
// 关键点：
//   - defer f.Close() 确保函数返回前关闭文件
//   - io.EOF 是正常的"读完"信号，不算错误
func ReadStream(path string, chunkSize int, onChunk func(chunk []byte) error) error {
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("打开文件 %q 失败：%w", path, err)
	}
	defer f.Close() // 即使后续 return error，文件也会被关闭

	if chunkSize <= 0 {
		chunkSize = 4096 // 默认4KB
	}
	buf := make([]byte, chunkSize)

	for {
		// n: 实际读取到的字节数 （可能小于缓冲区大小）
		n, err := f.Read(buf)

		// 先处理已读到的字节（即使后面有 EOF 也要先处理）
		if n > 0 {
			if cbErr := onChunk(buf[:n]); cbErr != nil {
				return cbErr
			}
		}

		// 然后判断 err
		if err == io.EOF {
			return nil // 正常读完
		}
		if err != nil {
			return fmt.Errorf("读取 %q 时出错：%w", path, err)
		}
	}
}
