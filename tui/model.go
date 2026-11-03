package tui

import (
	"MysqlAgent/core"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
)

// 全局渲染器, 宽度匹配视口
var mdRenderer, _ = glamour.NewTermRenderer(
	glamour.WithAutoStyle(),
	glamour.WithWordWrap(80),
)

// 问题样式: 右对齐 蓝色
var questionStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("231")). // 文字: 近白
	Background(lipgloss.Color("27")).  // 底色: 深蓝
	Padding(0, 1)                      // 左右留一点点内边距, 别贴边

// model状态
type model struct {
	input          textinput.Model // 输入框组件
	viewport       viewport.Model  // 可滚动窗口->显示内容
	loading        bool            // 模型是否正在处理
	history        []string        // 历史对话记录
	session        *core.Session   // 会话
	spinner        spinner.Model   // 加载动画
	scrollToBottom bool            // 是否需要滑动到底部
	mcpServer      string
}

// 输入框的边框样式
var inputBoxStyle = lipgloss.NewStyle().
	Border(lipgloss.RoundedBorder()).
	Padding(0, 1).
	BorderForeground(lipgloss.Color("76"))

// agent 返回的结果
type agentResultMsg struct {
	text string
}

// initialModel
func InitialModel(serverCommand string) model {
	// input
	// 创建一个输入框
	it := textinput.New()
	// 输入框为空,显示“请输入问题...”
	it.Placeholder = "请输入问题..."
	// 创建焦点来打字
	it.Focus()
	// viewport
	// 滚动区大小,初始内容为空
	vp := viewport.New(80, 20)
	vp.SetContent("")
	// spinner加载动画,蓝色加载点
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().
		Foreground(lipgloss.Color("27"))
	return model{
		input:     it,
		viewport:  vp,
		loading:   false,
		session:   core.NewAgentSession(),
		spinner:   sp,
		mcpServer: serverCommand,
	}
}

// Init
func (m model) Init() tea.Cmd {
	return nil
}

// 自定义 tick 消息, 驱动 spinner 动画
type tickMsg time.Time

// Update
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case spinner.TickMsg:
		if m.loading {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}
	case tea.KeyMsg:
		// 把按键交给输入框,让输入框接收字符
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)

		// 判断动作种类
		switch msg.String() {
		// 退出
		case "q", "ctrl+c":
			return m, tea.Quit
		// 翻页/滚动键交给 viewport 处理(Windows 终端下鼠标滚轮不可靠,用键盘替代)
		case "pgup", "pgdown", "up", "down":
			var vcmd tea.Cmd
			m.viewport, vcmd = m.viewport.Update(msg)
			return m, vcmd
		// 输入问题后回车
		case "enter":
			// 取到用户输入的问题
			question := m.input.Value()
			// 若输入为空,不处理
			if question == "" {
				return m, cmd
			}
			// 清空输入框
			m.input.SetValue("")
			// 将用户问题 右对齐 渲染
			q := questionStyle.Render(question)
			// 将问题追加到历史记录
			m.history = append(m.history, q)
			// 新内容来了, 标记要滚到底
			m.scrollToBottom = true
			// 历史记录+刚才的问题 放到可视区viewport
			m.viewport.SetContent(strings.Join(m.history, "\n"))
			// 将可视区滚动到底部
			m.viewport.GotoBottom()
			// loading改为true,正在加载
			m.loading = true
			// 返回两个 cmd: 一个是 Agent 异步任务, 一个是 spinner 动画 tick
			return m, tea.Batch(
				m.spinner.Tick, // ← 让加载图标开始转
				func() tea.Msg { // ← 原来的 AgentRun（异步）
					ans := core.AgentRun(m.session, question, m.mcpServer)
					return agentResultMsg{text: ans}
				},
			)
		default:
			return m, cmd
		}
	case tea.WindowSizeMsg:
		// 可视区随着windows终端窗口的大小变化
		m.viewport.Width = msg.Width
		m.viewport.Height = msg.Height
		return m, nil
	case tea.MouseMsg:
		// 鼠标滑轮滚动,翻看上下文
		var cmd tea.Cmd
		m.viewport, cmd = m.viewport.Update(msg)
		return m, cmd
	case agentResultMsg:
		// 修改loading
		m.loading = false
		// 收到Agent结果,把答案渲染后 追加到可视区
		ans, _ := mdRenderer.Render(msg.text)
		// 追加答案到history
		m.history = append(m.history, ans)
		// 答案来了, 滚到底看
		m.scrollToBottom = true
		// 历史记录+刚才的答案 放到可视区viewport
		m.viewport.SetContent(strings.Join(m.history, "\n"))
		// 自动滑到底部看答案
		m.viewport.GotoBottom()
		return m, nil
	}

	return m, nil
}

// View
func (m model) View() string {

	content := strings.Join(m.history, "\n")
	if m.loading {
		content += "\n" + m.spinner.View() + " 思考中..."
	}
	m.viewport.SetContent(content)
	if m.scrollToBottom {
		m.viewport.GotoBottom()
		m.scrollToBottom = false // 只滚一次, 立即重置
	}
	vp := m.viewport.View()
	input := m.input.View()
	// 将输入框打包成边框样式
	input = inputBoxStyle.Render(input)
	// 排版
	return lipgloss.JoinVertical(
		lipgloss.Left, // 靠左对齐
		vp,            //历史对话
		"",            // 空行,作为上下文间距
		input,         // 底部输入框
		"",
	)
}
