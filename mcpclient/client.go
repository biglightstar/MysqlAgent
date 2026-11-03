package mcpclient

import (
	"context"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/responses"
)

type Client struct {
	c     *client.Client
	tools []mcp.Tool
}

// New 连接到某个 stdio 方式的 MCP Server
func New(serverCommand string) (*Client, error) {
	// 拉起server子进程,建立stdio通信通道
	c, err := client.NewStdioMCPClient(serverCommand, nil)
	if err != nil {
		return nil, fmt.Errorf("启动 MCP Server 失败:%W", err)
	}
	// 握手:告诉server 所用的协议版本 客户端的身份信息
	initReq := mcp.InitializeRequest{}
	initReq.Params.ProtocolVersion = mcp.LATEST_LEGACY_PROTOCOL_VERSION // 最新协议
	initReq.Params.ClientInfo = mcp.Implementation{
		Name:    "MysqlAgent",
		Version: "1.0.0",
	}
	if _, err := c.Initialize(context.Background(), initReq); err != nil {
		c.Close() // 关闭子进程
		return nil, fmt.Errorf("MCP握手失败:%w", err)
	}
	// 列举sever提供的所有工具 缓存起来以后使用
	toolsResult, err := c.ListTools(context.Background(), mcp.ListToolsRequest{})
	if err != nil {
		c.Close()
		return nil, fmt.Errorf("列举工具失败:%w", err)
	}
	return &Client{c: c, tools: toolsResult.Tools}, nil
}

// 结束子进程
func (c *Client) Close() error {
	return c.c.Close()
}

// 调用指定名字的工具 返回字符串结果
func (c *Client) CallTool(ctx context.Context, name string, args map[string]any) (string, error) {
	// 调用工具
	result, err := c.c.CallTool(ctx, mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      name,
			Arguments: args,
		},
	})
	if err != nil {
		return "", err
	}
	// result.Content 是一组"内容项"(Content 接口),可能是文本、图片、音频等。
	// 我们只关心文本型的,用类型断言把它从接口里还原成 TextContent,再取 .Text
	var sb strings.Builder
	for _, content := range result.Content {
		if text, ok := content.(mcp.TextContent); ok {
			sb.WriteString(text.Text)
		}
	}
	return sb.String(), nil

}

// 把工具列表转化成openai认识的格式
func ToOpenAITools(tools []mcp.Tool) []responses.ToolUnionParam {
	result := make([]responses.ToolUnionParam, 0, len(tools))
	for _, t := range tools {
		result = append(result, responses.ToolUnionParam{
			OfFunction: &responses.FunctionToolParam{
				Name:        t.Name,
				Description: openai.String(t.Description),
				Parameters:  toOpenAIParameters(t.InputSchema),
			},
		})
	}
	return result
}

func toOpenAIParameters(schema mcp.ToolInputSchema) openai.FunctionParameters {
	return openai.FunctionParameters{
		"type":       schema.Type,
		"properties": schema.Properties,
		"required":   schema.Required,
	}
}

func (c *Client) Tools() []mcp.Tool {
	return c.tools
}
