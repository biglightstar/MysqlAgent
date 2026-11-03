# MysqlAgent

一个基于 **MCP（Model Context Protocol）** 架构的智能体，让你用自然语言操作 MySQL 数据库——想问什么直接问，想改数据直接说。

## 项目简介

MysqlAgent 把「大模型智能体」和「数据库操作」拆成两个独立的部分：

- **MCP Server**：独立的本地工具箱，提供 `SqlSelect`（查询）和 `SqlWrite`（写入）两个工具，封装对 MySQL 的操作。
- **MCP Client + Agent**：主程序作为客户端，通过标准输入输出（stdio）连接 Server，由大模型自动判断该调用哪个工具、传什么参数。

好处是客户端与服务端完全解耦：以后把「自己的 Server」换成「别人的 Server」，客户端代码基本不用改，只要改连接目标。

## 功能特性

- 自然语言查询数据库数据（SqlSelect）
- 自然语言写入 / 修改数据（SqlWrite）
- 基于 MCP 协议，客户端与服务端解耦
- 终端交互界面（TUI），支持对话式问答

## 技术栈

- **Go**：主要开发语言
- **mark3labs/mcp-go**：MCP 协议的 Go 实现（社区版）
- **openai-go**：调用大模型 API（兼容 OpenAI 协议，这里使用 DeepSeek）
- **bubbletea + bubbles**：终端界面（TUI）框架
- **MySQL**：目标数据库

## 环境要求

- Go 1.x（以 `go version` 输出为准）
- 一个可连接的 MySQL 实例
- 一个大模型 API Key（使用 DeepSeek，也可换成其他 OpenAI 兼容接口）

## 快速开始

### 1. 编译 MCP Server

Server 是独立子进程，需要先编译成可执行文件：

```bash
go build -o server.exe ./cmd/server
```

> 注意：主程序以子进程方式启动 `server.exe`，这一步不能少。

### 2. 运行主程序

```bash
go run .
```

启动后进入终端交互界面，直接输入问题，例如：

```
查询所有用户表里注册时间最近的前 10 个用户
```

## 项目结构

```
├── main.go          # 主入口（MCP Client + TUI）
├── cmd/server/      # MCP Server 独立入口（子进程）
├── mcpclient/       # MCP 客户端封装
├── mcpserver/       # MCP 服务端（注册 SqlSelect / SqlWrite）
├── tools/           # 具体工具实现（SQL 逻辑）
├── core/            # Agent 核心逻辑（会话、循环调用）
└── tui/             # 终端界面
```

## 配置说明

本项目通过**环境变量**读取配置信息，其中必填项需运行前设置，可选项有默认值、不设置也能跑。

### 必填

- **MODEL_API_KEY**：大模型 API Key（DeepSeek）
- **MYSQL_DSN**：MySQL 连接串，格式 `用户名:密码@tcp(主机:端口)/`

### 可选（不设置则使用默认值）

- **MODEL_BASE_URL**：大模型服务地址，默认 `https://api.deepseek.com`（换其他 OpenAI 兼容平台时设置）
- **MODEL_NAME**：模型名，默认 `deepseek-flash`

### Windows（PowerShell）设置示例

```powershell
$env:MODEL_API_KEY = "你的key"
$env:MYSQL_DSN = "root:你的密码@tcp(localhost:3306)/"
```

### Linux / macOS 设置示例

```bash
export MODEL_API_KEY="你的key"
export MYSQL_DSN="root:你的密码@tcp(localhost:3306)/"
```

> 注意：以上命令只对当前终端会话有效，关闭后需重新设置；也可写入系统的永久环境变量。

## 许可证

本项目采用 [MIT License](LICENSE) 开源，可自由使用、修改与分发，但须保留原作者版权声明

