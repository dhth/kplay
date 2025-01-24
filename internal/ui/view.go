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
	var footer string
	var msgsViewPtr string
	var mode string
	var msgMetadataTitleStyle lipgloss.Style
	var msgValueTitleStyle lipgloss.Style

	m.kMsgsList.Styles.Title = m.kMsgsList.Styles.Title.Background(lipgloss.Color(inactivePaneColor))
	msgMetadataTitleStyle = msgDetailsTitleStyle
	msgValueTitleStyle = msgDetailsTitleStyle

	switch m.activeView {
	case kMsgsListView:
		m.kMsgsList.Styles.Title = m.kMsgsList.Styles.Title.Background(lipgloss.Color(activeHeaderColor))
	case kMsgMetadataView:
		msgMetadataTitleStyle = msgMetadataTitleStyle.Background(lipgloss.Color(activeHeaderColor))
	case kMsgValueView:
		msgValueTitleStyle = msgValueTitleStyle.Background(lipgloss.Color(activeHeaderColor))
	}

	if m.persistRecords {
		mode += " " + persistingStyle.Render("persisting records!")
	}

	if m.skipRecords {
		mode += " " + skippingStyle.Render("skipping records!")
	}

	m.kMsgsList.Title += msgsViewPtr

	var statusBar string
	if m.msg != "" {
		statusBar = utils.TrimRight(m.msg, 120)
	}
	var errorMsg string
	if m.errorMsg != "" {
		errorMsg = " error: " + utils.TrimRight(m.errorMsg, 120)
	}

	var msgMetadataVPContent string
	if !m.msgValueVPReady {
		msgMetadataVPContent = vpNotReadyMsg
	} else {
		msgMetadataVPContent = fmt.Sprintf("%s\n\n%s\n", msgMetadataTitleStyle.Render("Message Metadata"), m.msgMetadataVP.View())
	}

	var msgValueVPContent string
	if !m.msgValueVPReady {
		msgValueVPContent = vpNotReadyMsg
	} else {
		msgValueVPContent = fmt.Sprintf("%s\n\n%s\n", msgValueTitleStyle.Render("Message Value"), m.msgValueVP.View())
	}
	var helpVPContent string
	if !m.helpVPReady {
		helpVPContent = vpNotReadyMsg
	} else {
		helpVPContent = fmt.Sprintf("%s\n\n%s\n", helpVPTitleStyle.Render("Help"), m.helpVP.View())
	}

	switch m.vpFullScreen {
	case false:
		content = lipgloss.JoinHorizontal(
			lipgloss.Top,
			messageListStyle.Render(m.kMsgsList.View()),
			lipgloss.JoinVertical(lipgloss.Left,
				viewPortStyle.Render(msgMetadataVPContent),
				viewPortStyle.Render(msgValueVPContent),
			),
		)
	case true:
		switch m.activeView {
		case kMsgMetadataView:
			content = viewPortFullScreenStyle.Render(msgMetadataVPContent)
		case kMsgValueView:
			content = viewPortFullScreenStyle.Render(msgValueVPContent)
		case helpView:
			content = viewPortFullScreenStyle.Render(helpVPContent)
		}
	}

	footerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#282828")).
		Background(lipgloss.Color("#7c6f64"))

	var helpMsg string
	if m.showHelpIndicator {
		helpMsg = " " + helpMsgStyle.Render("Press ? for help")
	}
	kConfigMsg := kConfigStyle.Render(fmt.Sprintf(" [%s] ", utils.TrimLeft(m.config.Topic, 40)))

	footerStr := fmt.Sprintf("%s%s%s%s%s",
		modeStyle.Render("kplay"),
		kConfigMsg,
		helpMsg,
		mode,
		errorMsg,
	)
	footer = footerStyle.Render(footerStr)

	return lipgloss.JoinVertical(lipgloss.Left,
		content,
		statusBar,
		footer,
	)
}
