// netsvc/http_client.go
//
// HTTP 客户端封装：GET / POST / 自定义 Header 等常用操作。
//
// 设计：使用自定义 *http.Client（带超时），避免 http.DefaultClient 永不超时的坑。
package netsvc

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"time"
)

// HTTPClient 是带默认超时的 HTTP 客户端。
type HTTPClient struct {
	client *http.Client
}

// NewHTTPClient 创建一个 HTTP 客户端。
//
// timeout 是整个请求（包括连接建立、TLS 握手、读响应体）的总超时。
// 推荐值：5-30 秒。设为 0 表示不超时（不推荐，会泄漏 goroutine）。
func NewHTTPClient(timeout time.Duration) *HTTPClient {
	return &HTTPClient{
		client: &http.Client{
			Timeout: timeout,
		},
	}
}

// Get 发起 GET 请求，返回响应体字节。
//
// 调用方拿到 body 后无需关心 resp.Body.Close（这里已经处理）。
func (c *HTTPClient) Get(url string) (statusCode int, body []byte, err error) {
	resp, err := c.client.Get(url)

	if err != nil {
		return 0, nil, fmt.Errorf("GET %q 失败: %w", url, err)
	}

	defer resp.Body.Close()

	body, err = io.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, nil, fmt.Errorf("读取响应体失败: %w", err)
	}
	return resp.StatusCode, body, nil
}

// Post 发起 POST 请求。
//
// contentType 常见值：
//
//	"application/json"
//	"application/x-www-form-urlencoded"
//	"text/plain"
func (c *HTTPClient) Post(url, contentType string, body []byte) (int, []byte, error) {
	resp, err := c.client.Post(url, contentType, bytes.NewReader(body))
	if err != nil {
		return 0, nil, fmt.Errorf("POST %q 失败: %w", url, err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, nil, fmt.Errorf("读取响应体失败: %w", err)
	}
	return resp.StatusCode, respBody, nil
}

// Head 发起 HEAD 请求，仅获取响应头（用于检查资源是否存在/已修改）。
//
// 比 GET 更快，不下载 body。健康检查模块（第6.1）会用到。
func (c *HTTPClient) Head(url string) (statusCode int, headers http.Header, err error) {
	resp, err := c.client.Head(url)
	if err != nil {
		return 0, nil, fmt.Errorf("HEAD %q 失败: %w", url, err)
	}
	defer resp.Body.Close()
	return resp.StatusCode, resp.Header, nil
}
