package core

import (
	"MysqlAgent/mcpclient"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/openai/openai-go/v3/responses"
)

// // 获取工具
// func getLlmTools() map[string]*tools.ToolDef {
// 	// 定义工具箱
// 	llmTools := make(map[string]*tools.ToolDef)
// 	// 调用函数创建工具: 读、写
// 	sqlSelectTool := tools.NewSqlSelect()
// 	sqlWriteTool := tools.NewSqlWrite()
// 	// 把工具放进工具箱
// 	llmTools[sqlSelectTool.Name] = sqlSelectTool
// 	llmTools[sqlWriteTool.Name] = sqlWriteTool
// 	// 返回结果
// 	return llmTools
// }

// // 将工具转化为openai认识的格式
// func buildOpenaiTools(tools map[string]*tools.ToolDef) []responses.ToolUnionParam {
// 	openaiTools := make([]responses.ToolUnionParam, 0)
// 	// 依次将工具转换openai格式工具
// 	for _, tool := range tools {
// 		openaiTools = append(openaiTools, tool.ToParam())
// 	}
// 	return openaiTools
// }

// agentRun函数
// 拿到会话session,运行agent
func AgentRun(session *Session, question string, serverCommand string) string {
	// 结果
	var ans string
	// 系统设置
	session.AddMessage(
		responses.ResponseInputItemParamOfMessage(
			question,
			responses.EasyInputMessageRoleUser),
	)
	// 构建上下文对象
	ctx := context.Background()
	// 连接mcp server  动态拉取工具列表
	mcpClient, err := mcpclient.New(serverCommand)
	if err != nil {
		return "连接mcp失败:" + err.Error()
	}
	defer mcpClient.Close()
	// 把mcp工具转成deep seek认识的openai模式
	openaiTools := mcpclient.ToOpenAITools(mcpClient.Tools())

	// // 获取openai式工具箱
	// llmtools := getLlmTools()
	// openaiTools := buildOpenaiTools(llmtools)

	// 读模型服务地址和模型名,没读到就用默认值(方便别人不改代码直接跑)
	baseURL := os.Getenv("MODEL_BASE_URL")
	if baseURL == "" {
		baseURL = "https://api.deepseek.com"
	}
	modelName := os.Getenv("MODEL_NAME")
	if modelName == "" {
		modelName = "deepseek-flash"
	}

	// 创建client对象
	client := openai.NewClient(
		option.WithAPIKey(os.Getenv("MODEL_API_KEY")),
		option.WithBaseURL(baseURL),
		option.WithRequestTimeout(30*time.Second),
	)
	// 开始循环: 感知-决策-执行
	var MaxLoop = 11
	for realLoop := 0; realLoop < MaxLoop; realLoop++ {
		// 向模型发送请求
		resp, err := client.Responses.New(ctx, responses.ResponseNewParams{
			Input: responses.ResponseNewParamsInputUnion{
				OfInputItemList: session.GetMessages(),
			},
			Tools: openaiTools,
			Model: modelName,
		})
		// 若有错误
		if err != nil {
			var apiErr *openai.Error
			if errors.As(err, &apiErr) {
				fmt.Println("HTTP 状态码:", apiErr.StatusCode)
				fmt.Println("错误类型:", apiErr.Type)
				fmt.Println("原始响应体:", apiErr.RawJSON())
			} else {
				fmt.Println("非 API 错误:", err)
			}
			panic(err)
		}
		// 初始化: 没有调用工具
		hasTool := false
		// 处理模型返回的结果
		for _, item := range resp.Output {
			// 判断每一条结果的类型
			if item.Type == "function_call" {
				// 调用了工具
				hasTool = true
				// 感知---记录信息到session
				session.AddMessage(responses.ResponseInputItemUnionParam{
					OfFunctionCall: &responses.ResponseFunctionToolCallParam{
						Arguments: item.AsFunctionCall().Arguments, // 对应工具需要的参数
						CallID:    item.AsFunctionCall().CallID,
						Name:      item.AsFunctionCall().Name, // 要调用的工具名
					},
				})
				// 把工具参数从json字符串解析成map[string]any
				var args map[string]any
				if err := json.Unmarshal([]byte(item.AsFunctionCall().Arguments), &args); err != nil {
					args = map[string]any{}
				}
				// 执行: 通过MCP协议调用工具
				toolRe, err := mcpClient.CallTool(ctx, item.AsFunctionCall().Name, args)
				if err != nil {
					toolRe = fmt.Sprintf("%s%v", item.AsFunctionCall().Name, err)
				}

				// // 开始调用工具
				// if v, ok := llmtools[item.AsFunctionCall().Name]; ok {
				// 	// 调用工具的结果
				// 	toolRe, err := v.Call(ctx, item.AsFunctionCall().Arguments)
				// 	if err != nil {
				// 		toolRe = fmt.Sprintf("%s%v", item.AsFunctionCall().Name, err)
				// 	}
				// 感知---记录工具调用的结果到session
				session.AddMessage(responses.ResponseInputItemUnionParam{
					OfFunctionCallOutput: &responses.ResponseInputItemFunctionCallOutputParam{
						Output: responses.ResponseInputItemFunctionCallOutputOutputUnionParam{
							OfString: openai.String(toolRe),
						},
						CallID: openai.String(item.AsFunctionCall().CallID),
					},
				})
			}
		}
		// 如果没有调用工具,返回答案
		if !hasTool {
			ans = resp.OutputText()
		}
	}
	// 感知---把最终答案作为assistant消息存进session
	if ans != "" {
		session.AddMessage(
			responses.ResponseInputItemParamOfMessage(
				ans,
				responses.EasyInputMessageRoleAssistant,
			),
		)
	}
	return ans
}
