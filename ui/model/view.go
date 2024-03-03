package model

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

var (
	listWidth = 50
)

func (m model) View() string {
	var content string
	var footer string
	var mode string
	var msgsViewPtr string
	var headerViewPtr string
	var valueViewPtr string

	switch m.activeView {
	case kMsgsListView:
		mode = "MESSAGES"
		msgsViewPtr = " 👇"
	case kMsgHeaderView:
		mode = "HEADERS"
		headerViewPtr = " 👇"
	case kMsgValueView:
		mode = "VALUE"
		valueViewPtr = " 👇"
	}

	m.kMsgsList.Title += msgsViewPtr

	var statusBar string
	if m.msg != "" {
		statusBar = Trim(m.msg, 120)
	}
	var errorMsg string
	if m.errorMsg != "" {
		errorMsg = "error: " + Trim(m.errorMsg, 120)
	}

	var msgHeadersVP string
	if !m.msgValueVPReady {
		msgHeadersVP = "\n  Initializing..."
	} else {
		msgHeadersVP = viewPortStyle.Render(fmt.Sprintf("%s%s\n\n%s\n", kMsgHeadersTitleStyle.Render("Message Headers"), headerViewPtr, m.msgHeadersVP.View()))
	}

	var msgValueVP string
	if !m.msgValueVPReady {
		msgValueVP = "\n  Initializing..."
	} else {
		msgValueVP = viewPortStyle.Render(fmt.Sprintf("%s%s\n\n%s\n", kMsgValueTitleStyle.Render("Message Value"), valueViewPtr, m.msgValueVP.View()))
	}

	switch m.vpFullScreen {
	case false:
		content = lipgloss.JoinHorizontal(
			lipgloss.Top,
			stackListStyle.Render(m.kMsgsList.View()),
			lipgloss.JoinVertical(lipgloss.Left,
				msgHeadersVP,
				msgValueVP,
			),
		)
	case true:
		switch m.activeView {
		case kMsgHeaderView:
			content = msgHeadersVP
		case kMsgValueView:
			content = msgValueVP
		}
	}

	footerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#282828")).
		Background(lipgloss.Color("#7c6f64"))

	footerStr := fmt.Sprintf("%s %s %s",
		modeStyle.Render(mode),
		"kplay",
		errorMsg,
	)
	footer = footerStyle.Render(footerStr)

	return lipgloss.JoinVertical(lipgloss.Left,
		content,
		statusBar,
		footer,
	)
}
