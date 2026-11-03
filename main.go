package main

import (
	"MysqlAgent/tui"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	// 新建program
	p := tea.NewProgram(
		tui.InitialModel("./server"),
		tea.WithMouseAllMotion(), // 用 all motion 模式,在 Windows 终端下滚轮兼容性更好
		tea.WithAltScreen(),      // 替代屏模式
	)
	// p.run
	if _, err := p.Run(); err != nil {
		fmt.Println("程序出错:", err)
	}
}
