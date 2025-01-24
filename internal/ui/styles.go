package ui

import "github.com/charmbracelet/lipgloss"

const (
	defaultBackgroundColor   = "#282828"
	defaultForegroundColor   = "#ebdbb2"
	listColor                = "#fe8019"
	activeHeaderColor        = "#fe8019"
	inactivePaneColor        = "#bdae93"
	listPaneBorderColor      = "#363230"
	topicColor               = "#d3869b"
	helpMsgColor             = "#83a598"
	helpViewTitleColor       = "#83a598"
	helpHeaderColor          = "#83a598"
	helpSectionColor         = "#fabd2f"
	successMsgcolor          = "#83a598"
	errorMsgcolor            = "#fb4934"
	toolNameColor            = "#b8bb26"
	persistingMsgsColor      = "#fb4934"
	skippingMsgsColor        = "#fabd2f"
	msgDetailsHeadingColor   = "#fabd2f"
	msgDetailsTombstoneColor = "#a89984"
)

var (
	baseStyle = lipgloss.NewStyle().
			PaddingLeft(1).
			PaddingRight(1).
			Foreground(lipgloss.Color(defaultBackgroundColor))

	baseListStyle = lipgloss.
			NewStyle().
			PaddingTop(1)

	messageListStyle = baseListStyle.
				Width(listWidth).
				Border(lipgloss.NormalBorder(), false, true, false, false).
				BorderForeground(lipgloss.Color(listPaneBorderColor))

	viewPortStyle = lipgloss.
			NewStyle().
			PaddingTop(1).
			PaddingLeft(2)

	viewPortFullScreenStyle = baseListStyle.
				PaddingLeft(2)

	toolNameStyle = baseStyle.
			Align(lipgloss.Center).
			Bold(true).
			Background(lipgloss.Color(toolNameColor))

	msgDetailsTitleStyle = baseStyle.
				Bold(true).
				Background(lipgloss.Color(inactivePaneColor)).
				Align(lipgloss.Left)

	topicStyle = baseStyle.
			Bold(true).
			Foreground(lipgloss.Color(topicColor))

	persistingStyle = baseStyle.
			Bold(true).
			Foreground(lipgloss.Color(persistingMsgsColor))

	skippingStyle = baseStyle.
			Bold(true).
			Foreground(lipgloss.Color(skippingMsgsColor))

	helpMsgStyle = baseStyle.
			Bold(true).
			Foreground(lipgloss.Color(helpMsgColor))

	helpVPTitleStyle = baseStyle.
				Bold(true).
				Background(lipgloss.Color(helpViewTitleColor)).
				Align(lipgloss.Left)

	helpHeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(helpHeaderColor))

	helpSectionStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(helpSectionColor))

	successMsgStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(successMsgcolor))

	errorMsgStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(errorMsgcolor))

	msgDetailsHeadingStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color(msgDetailsHeadingColor))

	msgDetailsErrorStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(errorMsgcolor))

	msgDetailsTombstoneStyle = lipgloss.NewStyle().
					PaddingLeft(1).
					PaddingRight(1).
					Foreground(lipgloss.Color(defaultBackgroundColor)).
					Background(lipgloss.Color(msgDetailsTombstoneColor))
)
