package ui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/dhth/kplay/internal/utils"
)

var (
	listWidth     = 44
	vpNotReadyMsg = "\n  Initializing..."
)

func (m Model) View() string {
	var content string
	var msgsViewPtr string
	var mode string
	var msgValueTitleStyle lipgloss.Style

	m.msgsList.Styles.Title = m.msgsList.Styles.Title.Background(lipgloss.Color(inactivePaneColor))
	msgValueTitleStyle = msgDetailsTitleStyle

	switch m.activeView {
	case kMsgsListView:
		m.msgsList.Styles.Title = m.msgsList.Styles.Title.Background(lipgloss.Color(activeHeaderColor))
	case kMsgValueView:
		msgValueTitleStyle = msgValueTitleStyle.Background(lipgloss.Color(activeHeaderColor))
	}

	if m.persistRecords {
		mode += " " + persistingStyle.Render("persisting records!")
	}

	if m.skipRecords {
		mode += " " + skippingStyle.Render("skipping records!")
	}

	m.msgsList.Title += msgsViewPtr

	var statusBar string
	if m.msg != "" && m.errorMsg != "" {
		statusBar = fmt.Sprintf("%s %s", successMsgStyle.Render(m.msg), errorMsgStyle.Render(m.errorMsg))
	} else if m.msg != "" {
		statusBar = successMsgStyle.Render(m.msg)
	} else {
		statusBar = errorMsgStyle.Render(m.errorMsg)
	}

	var msgValueVPContent string
	if !m.msgDetailsVPReady {
		msgValueVPContent = vpNotReadyMsg
	} else {
		msgValueVPContent = fmt.Sprintf("%s\n\n%s\n", msgValueTitleStyle.Render("Message Value"), m.msgDetailsVP.View())
	}
	var helpVPContent string
	if !m.helpVPReady {
		helpVPContent = vpNotReadyMsg
	} else {
		helpVPContent = fmt.Sprintf("%s\n\n%s\n", helpVPTitleStyle.Render("Help"), m.helpVP.View())
	}

	switch m.activeView {
	case kMsgsListView, kMsgValueView:
		content = lipgloss.JoinHorizontal(
			lipgloss.Top,
			messageListStyle.Render(m.msgsList.View()),
			viewPortStyle.Render(msgValueVPContent),
		)
	case helpView:
		content = viewPortFullScreenStyle.Render(helpVPContent)
	}

	var helpMsg string
	if m.showHelpIndicator {
		helpMsg = " " + helpMsgStyle.Render("Press ? for help")
	}
	topicMarker := topicStyle.Render(fmt.Sprintf(" [%s] ", utils.TrimLeft(m.config.Topic, 40)))

	footer := fmt.Sprintf("%s%s%s%s",
		toolNameStyle.Render("kplay"),
		topicMarker,
		helpMsg,
		mode,
	)

	return lipgloss.JoinVertical(lipgloss.Left,
		content,
		statusBar,
		footer,
	)
}
