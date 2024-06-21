package model

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

var (
	listWidth = 40
)

func (m model) View() string {
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
		statusBar = TrimRight(m.msg, 120)
	}
	var errorMsg string
	if m.errorMsg != "" {
		errorMsg = " error: " + TrimRight(m.errorMsg, 120)
	}

	var msgMetadataVP string
	if !m.msgValueVPReady {
		msgMetadataVP = "\n  Initializing..."
	} else {
		msgMetadataVP = viewPortStyle.Render(fmt.Sprintf("%s\n\n%s\n", msgMetadataTitleStyle.Render("Message Metadata"), m.msgMetadataVP.View()))
	}

	var msgValueVP string
	if !m.msgValueVPReady {
		msgValueVP = "\n  Initializing..."
	} else {
		msgValueVP = viewPortStyle.Render(fmt.Sprintf("%s\n\n%s\n", msgValueTitleStyle.Render("Message Value"), m.msgValueVP.View()))
	}
	var helpVP string
	if !m.helpVPReady {
		helpVP = "\n  Initializing..."
	} else {
		helpVP = helpVPStyle.Render(fmt.Sprintf("  %s\n\n%s\n", helpVPTitleStyle.Render("Help"), m.helpVP.View()))
	}

	switch m.vpFullScreen {
	case false:
		content = lipgloss.JoinHorizontal(
			lipgloss.Top,
			stackListStyle.Render(m.kMsgsList.View()),
			lipgloss.JoinVertical(lipgloss.Left,
				msgMetadataVP,
				msgValueVP,
			),
		)
	case true:
		switch m.activeView {
		case kMsgMetadataView:
			content = msgMetadataVP
		case kMsgValueView:
			content = msgValueVP
		case helpView:
			content = helpVP
		}
	}

	footerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#282828")).
		Background(lipgloss.Color("#7c6f64"))

	var helpMsg string
	if m.showHelpIndicator {
		helpMsg = " " + helpMsgStyle.Render("Press ? for help")
	}
	kConfigMsg := kConfigStyle.Render(fmt.Sprintf(" [%s] ", TrimLeft(m.kconfig.Topic, 40)))

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
