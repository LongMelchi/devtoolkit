// datastore/event.go
// 定义本项目运行数据的"事件"结构 —— 所有写入 datastore 的记录
// 都使用同一种 schema，便于后续统一查询、聚合与展示。
package datastore

import "time"

// Event 表示一条运行事件。
// 设计原则：
//   - 必填字段（Time/Module/Action/Status）保证可索引、可统计
//   - Payload 用 map[string]any 容纳模块特有的自由字段，
//     既保持灵活性，又让大部分使用者不需要为新字段改 struct
type Event struct {
	Time    time.Time      `json:"time"`              // 事件发生时刻（RFC3339）
	Module  string         `json:"module"`            // 触发模块：file/data/net/proc/healthcheck/...
	Action  string         `json:"action"`            // 动作：read/write/grep/http_get/exec/...
	Status  string         `json:"status"`            // 结果：ok / error
	Message string         `json:"message,omitempty"` // 一句话摘要（可选）
	Payload map[string]any `json:"payload,omitempty"` // 自由字段（可选）
}

// NewEvent 构造一条 Event，自动填充时间。
// 用法示例：
//   datastore.NewEvent("file", "read", "ok").
//       WithMessage("read go.mod").
//       WithPayload("size", 256)
func NewEvent(module, action, status string) *Event {
	return &Event{
		Time:    time.Now().UTC(), // UTC 避免时区混乱，展示时再转本地
		Module:  module,
		Action:  action,
		Status:  status,
		Payload: make(map[string]any),
	}
}

// WithMessage 链式设置 Message。
func (e *Event) WithMessage(msg string) *Event {
	e.Message = msg
	return e
}

// WithPayload 链式追加一个 key/value 到 Payload。
func (e *Event) WithPayload(key string, value any) *Event {
	if e.Payload == nil {
		e.Payload = make(map[string]any)
	}
	e.Payload[key] = value
	return e
}
