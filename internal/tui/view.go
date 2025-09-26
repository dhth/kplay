package tui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/dhth/kplay/internal/utils"
)

const (
	listWidth          = 44
	configKeyMaxLength = 50
	minHeightNeeded    = 30
	minWidthNeeded     = 90
)

var vpNotReadyMsg = "\n  Initializing..."

func (m Model) View() string {
	var content string
	var behavioursMsg string
	var msgDetailsTitleStyle lipgloss.Style

	switch m.activeView {
	case msgListView:
		m.msgsList.Styles.Title = m.msgsList.Styles.Title.Background(lipgloss.Color(activeHeaderColor))
		msgDetailsTitleStyle = inactiveMsgDetailsTitleStyle
	case msgDetailsView:
		m.msgsList.Styles.Title = m.msgsList.Styles.Title.Background(lipgloss.Color(inactivePaneColor))
		msgDetailsTitleStyle = inactiveMsgDetailsTitleStyle.Background(lipgloss.Color(activeHeaderColor))
	}

	if m.behaviours.PersistMessages {
		behavioursMsg += persistingStyle.Render("persisting messages!")
	}

	if m.behaviours.SkipMessages {
		behavioursMsg += skippingStyle.Render("skipping messages!")
	}

	var statusBar string
	if m.msg != "" && m.errorMsg != "" {
		statusBar = fmt.Sprintf("%s %s", successMsgStyle.Render(m.msg), errorMsgStyle.Render(m.errorMsg))
	} else if m.msg != "" {
		statusBar = successMsgStyle.Render(m.msg)
	} else {
		statusBar = errorMsgStyle.Render(m.errorMsg)
	}

	var msgDetailsVPContent string
	if !m.msgDetailsVPReady {
		msgDetailsVPContent = vpNotReadyMsg
	} else {
		msgDetailsVPContent = fmt.Sprintf("%s\n\n%s\n", msgDetailsTitleStyle.Render("Message Details"), m.msgDetailsVP.View())
	}
	var helpVPContent string
	if !m.helpVPReady {
		helpVPContent = vpNotReadyMsg
	} else {
		helpVPContent = fmt.Sprintf("%s\n\n%s\n", helpVPTitleStyle.Render("Help"), m.helpVP.View())
	}

	switch m.activeView {
	case msgListView, msgDetailsView:
		content = lipgloss.JoinHorizontal(
			lipgloss.Top,
			messageListStyle.Render(m.msgsList.View()),
			viewPortStyle.Render(msgDetailsVPContent),
		)
	case helpView:
		content = viewPortFullScreenStyle.Render(helpVPContent)
	case insufficientDimensionsView:
		return fmt.Sprintf(`
  Terminal size too small:
    Width = %d Height = %d

  Minimum dimensions needed:
    Width = %d Height = %d

  Press q/<ctrl+c>/<esc>
    to exit
`, m.terminalWidth, m.terminalHeight, minWidthNeeded, minHeightNeeded)
	}

	var helpMsg string
	if m.showHelpIndicator {
		helpMsg = helpMsgStyle.Render("Press ? for help")
	}

	topic := topicStyle.Render(utils.TrimLeft(m.config.Topic, configKeyMaxLength))

	footer := fmt.Sprintf("%s  %s%s%s",
		toolNameStyle.Render("kplay"),
		topic,
		behavioursMsg,
		helpMsg,
	)

	return lipgloss.JoinVertical(lipgloss.Left,
		content,
		statusBar,
		footer,
	)
}
