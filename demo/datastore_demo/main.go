// demo/datastore_demo/main.go
//
// 第3章演示：以最小代码展示 datastore 的写入与读取。
package main

import (
	"fmt"
	"log"

	"devtoolkit/datastore"
)

func main() {
	fmt.Println("=== 3.1 打开 Store（自动创建 ./data/demo-YYYY-MM-DD.jsonl）===")
	store, err := datastore.Open("./data", "demo")
	if err != nil {
		log.Fatal(err)
	}
	defer store.Close()
	fmt.Printf("已打开: %s\n", store.Path())

	fmt.Println("\n=== 3.2 追加 3 条 Event（O_APPEND 模式）===")
	events := []*datastore.Event{
		datastore.NewEvent("demo", "create", "ok").
			WithMessage("hello jsonl").
			WithPayload("size", 42),
		datastore.NewEvent("demo", "update", "ok").
			WithMessage("update item 1").
			WithPayload("delta", 10),
		datastore.NewEvent("demo", "delete", "error").
			WithMessage("not found").
			WithPayload("id", 999),
	}
	for _, e := range events {
		if err := store.Append(e); err != nil {
			log.Println("写入失败:", err)
		}
	}
	fmt.Println("写入完毕")

	fmt.Println("\n=== 3.3 流式读取并打印 ===")
	all, err := datastore.ReadAll(store.Path())
	if err != nil {
		log.Println("读取失败:", err)
		return
	}
	for i, e := range all {
		fmt.Printf("  #%d  [%s] %s.%s status=%s msg=%q payload=%v\n",
			i+1, e.Time.Format("15:04:05"),
			e.Module, e.Action, e.Status, e.Message, e.Payload)
	}

	fmt.Println("\n=== 3.4 按模块列出 ./data 下所有 jsonl ===")
	files, _ := datastore.ListFiles("./data", "demo")
	for _, f := range files {
		fmt.Println("  -", f)
	}
}
