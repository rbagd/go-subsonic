package ui

import "github.com/charmbracelet/lipgloss"

var (
	appStyle = lipgloss.NewStyle().Padding(1, 2)

	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFDF5")).
			Background(lipgloss.Color("#25A065")).
			Padding(0, 1)

	statusMessageStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#AEAFAD")).
				MarginTop(1)

	// Panel styles
	panelStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("240")).
			Padding(1)

	focusedPanelStyle = panelStyle.Copy().
				BorderForeground(lipgloss.Color("62"))

	// List styles
	selectedItemStyle = lipgloss.NewStyle().
			PaddingLeft(2).
			Foreground(lipgloss.Color("170"))
	
	itemStyle = lipgloss.NewStyle().
			PaddingLeft(2)

	faintStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))

	// Layout help
	helpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
)
