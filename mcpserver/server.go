// ===================建 Server、注册 SqlSelect/SqlWrite
package mcpserver

import (
	"MysqlAgent/tools"
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func NewServer() *server.MCPServer {
	// 创建server对象
	s := server.NewMCPServer(
		"MysqlAgent",
		"1.0.0",
		server.WithToolCapabilities(true),
	)
	// 注册“只读查询”工具sqlselect
	s.AddTool(
		mcp.NewTool(
			"SqlSelect",
			mcp.WithDescription(`执行只读的 SQL 查询(SELECT、SHOW、DESCRIBE 等)，并返回查询结果
			当用户想要查看、检索、列出数据库中的数据或结构时使用,例如:
			"有哪些数据库" "这个表里有什么数据" "看看这张表的结构"。
			收到这类需求时,你把用户的意思转换成合适的 SQL 语句,交给本工具执行
			本工具只能执行读取操作,禁止用于修改或删除数据,
			调用本工具得到查询结果后,请用带边框的 ASCII 表格形式返回给用户,让每一列对齐列头
			切记生成的 SQL 必须写全库名和表名,例如"SELECT * FROM test.user",不要省略库名`),
			mcp.WithString(
				"sqlsentence",
				mcp.Required(),
				mcp.Description("只读的 SQL 查询语句(SELECT、SHOW、DESCRIBE 等)"),
			),
		),
		// 处理器
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			// 检查请求语句中是否要调用sql查询工具
			sqlsentence := req.GetString("sqlsentence", "")
			if sqlsentence == "" {
				return mcp.NewToolResultError("缺少sqlsentence参数"), nil
			}
			// 复用tools包里面的只读查询函数
			result, err := tools.RunSelectQuery(ctx, sqlsentence)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			// 把查询结果返回给客户端
			return mcp.NewToolResultText(result), nil
		},
	)
	// 注册“写操作”工具sqlwrite
	s.AddTool(
		// 工具描述
		mcp.NewTool(
			"SqlWrite",
			mcp.WithDescription(`执行修改数据库的 SQL 语句(INSERT、UPDATE、DELETE 等),
			例如新增记录、修改现有数据、删除记录,
			当用户想要写入、更新或删除数据库中的数据时使用,
			本工具在执行前会向用户确认是否同意修改;只有用户明确同意后才真正执行,
			调用本工具得到查询结果后,请用带边框的 ASCII 表格形式返回给用户,让每一列对齐列头
			切记生成的 SQL 必须写全库名和表名,例如"SELECT * FROM test.user",不要省略库名`),
			mcp.WithString(
				"sqlsentence",
				mcp.Required(),
				mcp.Description("需要写权限的 SQL 语句(INSERT、UPDATE、DELETE 等写操作)"),
			),
		),
		// 工具函数
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			sqlsentence := req.GetString("sqlsentence", "")
			if sqlsentence == "" {
				return mcp.NewToolResultError("缺少 sqlsentence 参数"), nil
			}
			// 复用tools中的写操作函数
			result, err := tools.RunWriteQuery(ctx, sqlsentence)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return mcp.NewToolResultText(result), nil
		},
	)
	// 返回配置好的server
	return s
}
