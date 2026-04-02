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

	panelHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("252"))

	focusedPanelHeaderStyle = panelHeaderStyle.Copy().
				Foreground(lipgloss.Color("230"))

	// List styles
	selectedItemStyle = lipgloss.NewStyle().
				PaddingLeft(2).
				Foreground(lipgloss.Color("170"))

	selectedItemMetaStyle = selectedItemStyle.Copy().
				Faint(true)

	itemStyle = lipgloss.NewStyle().
			PaddingLeft(2)

	recentItemStyle = itemStyle.Copy().
			Bold(true).
			Foreground(lipgloss.Color("220"))

	selectedRecentItemStyle = selectedItemStyle.Copy().
				Bold(true).
				Foreground(lipgloss.Color("16")).
				Background(lipgloss.Color("220"))

	faintStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))

	unfocusedSelectedItemStyle = itemStyle.Copy().
					Foreground(lipgloss.Color("241"))

	unfocusedSelectedItemMetaStyle = unfocusedSelectedItemStyle.Copy().
					Faint(true)

	boldStyle = lipgloss.NewStyle().Bold(true)

	statBoxStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#25A065")).
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("240")).
			Padding(0, 1).
			MarginBottom(1)

	// Layout help
	helpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))

	quitConfirmStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFFDF5")).
				Background(lipgloss.Color("#FF5F87")).
				Padding(0, 1).
				Bold(true)
)
