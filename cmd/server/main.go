// ================调用 server.go 的构造 + ServeStdio 启动
package main

import (
	"MysqlAgent/mcpserver"
	"fmt"
	"os"

	"github.com/mark3labs/mcp-go/server"
)

func main() {
	// 拿到已经注册好 SqlSelect、SqlWrite 两个工具的 Server
	s := mcpserver.NewServer()
	// 启动stdio通信
	if err := server.ServeStdio(s); err != nil {
		fmt.Fprintln(os.Stderr, "mcp服务运行出错:", err)
		os.Exit(1)
	}
}
