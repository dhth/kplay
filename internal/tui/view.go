package tui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/dhth/kplay/internal/utils"
)

const (
	configKeyMaxLength = 50
)

var (
	listWidth     = 44
	vpNotReadyMsg = "\n  Initializing..."
)

func (m Model) View() string {
	var content string
	var msgsViewPtr string
	var mode string
	var msgDetailsTitleStyle lipgloss.Style

	switch m.activeView {
	case msgListView:
		m.msgsList.Styles.Title = m.msgsList.Styles.Title.Background(lipgloss.Color(activeHeaderColor))
		msgDetailsTitleStyle = inactiveMsgDetailsTitleStyle
	case msgDetailsView:
		m.msgsList.Styles.Title = m.msgsList.Styles.Title.Background(lipgloss.Color(inactivePaneColor))
		msgDetailsTitleStyle = inactiveMsgDetailsTitleStyle.Background(lipgloss.Color(activeHeaderColor))
	}

	if m.persistMessages {
		mode += " " + persistingStyle.Render("persisting records!")
	}

	if m.skipMessages {
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
	}

	var helpMsg string
	if m.showHelpIndicator {
		helpMsg = helpMsgStyle.Render("Press ? for help")
	}

	configMsg := fmt.Sprintf("%s<-- %s", consumerGroupStyle.Render(utils.TrimLeft(m.config.ConsumerGroup, configKeyMaxLength)), topicStyle.Render(utils.TrimLeft(m.config.Topic, configKeyMaxLength)))

	footer := fmt.Sprintf("%s  %s%s%s",
		toolNameStyle.Render("kplay"),
		configMsg,
		helpMsg,
		mode,
	)

	return lipgloss.JoinVertical(lipgloss.Left,
		content,
		statusBar,
		footer,
	)
}
