// config/env.go
// 全局配置：从环境变量读取参数，提供合理默认值。
// 设计原则：本包不依赖任何其他业务包，只标准库。
package config

import (
	"os"
	"path/filepath"
	"strconv"
)

// Config 是全项目共享的运行时配置。
// 设计思路：把所有"可由环境变量覆盖"的参数集中放在一个结构体里，
// 避免散落在各模块中各自调用 os.Getenv，便于测试和文档化。
type Config struct {
	Host     string // 网络服务监听地址（DEVTOOLKIT_HOST）
	Port     int    // 网络服务监听端口（DEVTOOLKIT_PORT）
	LogLevel string // 日志级别（DEVTOOLKIT_LOG）：debug/info/warn/error
	LogFile  string // 日志文件路径（DEVTOOLKIT_LOGFILE），空表示只打印到控制台
	DataDir  string // 运行数据目录（DEVTOOLKIT_DATA_DIR），默认 ./data
}

// Load 读取所有支持的环境变量，并返回填充好的 Config。
// 优先级：环境变量 > 默认值。
// 这是一个纯函数（仅读环境，不修改全局状态），便于测试。
func Load() *Config {
	return &Config{
		Host:     getEnvDefault("DEVTOOLKIT_HOST", "localhost"),
		Port:     getEnvIntDefault("DEVTOOLKIT_PORT", 8080),
		LogLevel: getEnvDefault("DEVTOOLKIT_LOG", "info"),
		LogFile:  os.Getenv("DEVTOOLKIT_LOGFILE"), // 空字符串即可
		DataDir:  getEnvDefault("DEVTOOLKIT_DATA_DIR", filepath.Join(".", "data")),
	}
}

// getEnvDefault 是内部辅助：若环境变量不存在则返回 fallback。
// 注意首字母小写表示包内私有（go 的导出规则：大写=公开，小写=私有）。
func getEnvDefault(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}

// getEnvIntDefault 把环境变量解析为 int，失败则用 fallback。
func getEnvIntDefault(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	if n, err := strconv.Atoi(v); err == nil {
		return n
	}
	return fallback // 解析失败也用默认值，永不崩溃
}
