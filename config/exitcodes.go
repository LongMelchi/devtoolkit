// 设计说明：把"环境变量读取"和"退出码常量"独立成 `config/` 包，
// 所有上层模块（cmd/healthcheck/backup 等）共用同一份配置。

// config/exitcodes.go
// 项目统一的退出码常量，避免魔法数字散落各处。
// 退出码语义参考 Linux Shell 约定：
//
//	0       = 成功
//	1       = 通用错误
//	2       = 命令行参数错误（misuse of shell builtin）
//	64-78   = sysexits.h 定义的标准错误码
//	126/127 = 命令找不到 / 不可执行
package config

const (
	ExitOK           = 0  // 成功
	ExitGenericError = 1  // 通用错误（未分类）
	ExitUsageError   = 2  // 命令行参数错误
	ExitIOError      = 64 // 文件读写错误
	ExitNetworkError = 65 // 网络连接错误
	ExitProcessError = 66 // 子进程执行失败
	ExitConfigError  = 78 // 配置错误
)
