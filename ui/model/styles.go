package model

import "github.com/charmbracelet/lipgloss"

const (
	defaultBackgroundColor = "#282828"
	listColor              = "#fe8019"
	activeHeaderColor      = "#fe8019"
	inactivePaneColor      = "#928374"
	listPaneBorderColor    = "#3c3836"
	helpMsgColor           = "#83a598"
	helpViewTitleColor     = "#83a598"
	helpHeaderColor        = "#83a598"
	helpSectionColor       = "#fabd2f"
)

var (
	baseStyle = lipgloss.NewStyle().
			PaddingLeft(1).
			PaddingRight(1).
			Foreground(lipgloss.Color("#282828"))

	baseListStyle = lipgloss.
			NewStyle().
			PaddingTop(1).
			PaddingRight(4).
			PaddingLeft(1).
			PaddingBottom(1)

	stackListStyle = baseListStyle.Copy().
			Width(listWidth+10).
			Border(lipgloss.NormalBorder(), false, true, false, false).
			BorderForeground(lipgloss.Color(listPaneBorderColor))

	viewPortStyle = baseListStyle.
			Copy().
			Width(150).
			PaddingLeft(3)

	modeStyle = baseStyle.Copy().
			Align(lipgloss.Center).
			Bold(true).
			Background(lipgloss.Color("#b8bb26"))

	msgDetailsTitleStyle = baseStyle.Copy().
				Bold(true).
				Background(lipgloss.Color(inactivePaneColor)).
				Align(lipgloss.Left)

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
			Foreground(lipgloss.Color(helpMsgColor))

	helpVPTitleStyle = baseStyle.Copy().
				Bold(true).
				Background(lipgloss.Color(helpViewTitleColor)).
				Align(lipgloss.Left)

	helpHeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(helpHeaderColor))

	helpSectionStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(helpSectionColor))

	helpVPStyle = lipgloss.NewStyle().
			PaddingTop(1).
			PaddingRight(2).
			PaddingLeft(1).
			PaddingBottom(1)
)
