// netsvc/tcp_client.go
//
// TCP 客户端：连接 TCP 服务端，可读写多次后关闭。
package netsvc

import (
	"bufio"
	"fmt"
	"net"
	"time"
)

// TCPClient 是 TCP 客户端的简单封装。
type TCPClient struct {
	conn   net.Conn
	reader *bufio.Reader
}

// DialTCP 连接到 host:port。
//
// 通过 net.DialTimeout 而非 net.Dial，避免连接超时无限等待。
func DialTCP(host, port string, timeout time.Duration) (*TCPClient, error) {
	addr := host + ":" + port
	conn, err := net.DialTimeout("tcp", addr, timeout)
	if err != nil {
		return nil, fmt.Errorf("连接 %q 失败: %w", addr, err)
	}
	return &TCPClient{
		conn:   conn,
		reader: bufio.NewReader(conn),
	}, nil
}

// SendLine 发送一行文本（自动追加 \n）并返回服务端的一行响应。
func (c *TCPClient) SendLine(line string) (string, error) {
	if _, err := fmt.Fprintln(c.conn, line); err != nil {
		return "", fmt.Errorf("发送失败: %w", err)
	}
	resp, err := c.reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("接收失败: %w", err)
	}
	return resp, nil
}

// Close 关闭连接。
func (c *TCPClient) Close() error {
	return c.conn.Close()
}
