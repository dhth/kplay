package model

import "github.com/charmbracelet/lipgloss"

var (
	baseStyle = lipgloss.NewStyle().
			PaddingLeft(1).
			PaddingRight(1).
			Foreground(lipgloss.Color("#282828"))

	baseListStyle = lipgloss.NewStyle().PaddingTop(1).PaddingRight(4).PaddingLeft(1).PaddingBottom(1)

	stackListStyle = baseListStyle.Copy().Width(listWidth+10).Border(lipgloss.NormalBorder(), false, true, false, false).BorderForeground(lipgloss.Color("#343434"))
	viewPortStyle  = baseListStyle.Copy().Width(150).PaddingLeft(8)

	modeStyle = baseStyle.Copy().
			Align(lipgloss.Center).
			Bold(true).
			Background(lipgloss.Color("#b8bb26"))

	kMsgMetadataTitleStyle = baseStyle.Copy().
				Bold(true).
				Background(lipgloss.Color("#b8bb26")).
				Align(lipgloss.Left)

	kMsgValueTitleStyle = baseStyle.Copy().
				Bold(true).
				Background(lipgloss.Color("#8ec07c")).
				Align(lipgloss.Left)

	kConfigStyle = baseStyle.Copy().
			Bold(true).
			Foreground(lipgloss.Color("#d3869b"))

	persistingStyle = baseStyle.Copy().
			Bold(true).
			Foreground(lipgloss.Color("#fb4934"))

	skippingStyle = baseStyle.Copy().
			Bold(true).
			Foreground(lipgloss.Color("#fabd2f"))

	helpMsgStyle = baseStyle.Copy().
			Bold(true).
			Foreground(lipgloss.Color("#83a598"))
)
