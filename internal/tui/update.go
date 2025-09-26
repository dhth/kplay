package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	t "github.com/dhth/kplay/internal/types"
)

const (
	viewPortMoveLineCount = 3
	genericErrMsg         = "something went wrong"
	alreadyFetchingMsg    = "already fetching"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd
	forceRefreshMsgDetailsVP := false
	m.msg = ""
	m.errorMsg = ""

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.terminalHeight = msg.Height
		m.terminalWidth = msg.Width
		if msg.Width < minWidthNeeded || msg.Height < minHeightNeeded {
			if m.activeView != insufficientDimensionsView {
				m.lastViewBeforeInsufficientDims = m.activeView
				m.activeView = insufficientDimensionsView
			}
			return m, nil
		}

		if m.activeView == insufficientDimensionsView {
			m.activeView = m.lastViewBeforeInsufficientDims
		}

	case tea.KeyMsg:
		if m.activeView == insufficientDimensionsView {
			switch msg.String() {
			case "ctrl+c", "q", "esc":
				return m, tea.Quit
			default:
				return m, nil
			}
		}
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "q", "esc":
			switch m.activeView {
			case msgListView:
				return m, tea.Quit
			case msgDetailsView:
				m.activeView = msgListView
			case helpView:
				m.activeView = m.lastView
			}
		case "n", " ":
			if m.activeView == helpView {
				break
			}

			if m.fetchingInProgress {
				m.errorMsg = alreadyFetchingMsg
				break
			}

			cmds = append(cmds, FetchMessages(m.client, m.config, 1))
			m.fetchingInProgress = true
		case "N":
			if m.activeView == helpView {
				break
			}

			if m.fetchingInProgress {
				m.errorMsg = alreadyFetchingMsg
				break
			}

			cmds = append(cmds, FetchMessages(m.client, m.config, 10))
			m.fetchingInProgress = true
		case "}":
			if m.activeView == helpView {
				break
			}

			if m.fetchingInProgress {
				m.errorMsg = alreadyFetchingMsg
				break
			}

			cmds = append(cmds, FetchMessages(m.client, m.config, 100))
			m.fetchingInProgress = true
		case "?":
			if m.activeView != helpView {
				m.lastView = m.activeView
				m.activeView = helpView
			} else {
				m.activeView = m.lastView
			}
		case "p":
			if m.activeView == helpView {
				break
			}

			m.behaviours.PersistMessages = !m.behaviours.PersistMessages
		case "s":
			if m.activeView == helpView {
				break
			}

			m.behaviours.SkipMessages = !m.behaviours.SkipMessages
		case "y":
			if len(m.msgsList.Items()) == 0 {
				break
			}

			if m.activeView == helpView {
				break
			}

			message, ok := m.msgsList.SelectedItem().(t.Message)
			if !ok {
				m.errorMsg = genericErrMsg
				break
			}

			if !ok {
				break
			}

			detailsStr := message.GetDetails()
			cmds = append(cmds, copyToClipboard(detailsStr))
		case "[":
			if m.activeView == helpView {
				break
			}

			if len(m.msgsList.Items()) == 0 {
				break
			}

			if m.msgsList.Index() == 0 {
				break
			}

			m.msgsList.CursorUp()
		case "]":
			if m.activeView == helpView {
				break
			}

			if len(m.msgsList.Items()) == 0 {
				break
			}

			if m.msgsList.Index() == len(m.msgsList.Items())-1 {
				break
			}

			m.msgsList.CursorDown()
		case "j", "down":
			switch m.activeView {
			case msgListView:
				m.msgDetailsVP.GotoTop()
			case msgDetailsView:
				if m.msgDetailsVP.AtBottom() {
					break
				}
				m.msgDetailsVP.ScrollDown(viewPortMoveLineCount)
			case helpView:
				if m.helpVP.AtBottom() {
					break
				}
				m.helpVP.ScrollDown(viewPortMoveLineCount)
			}
		case "k", "up":
			switch m.activeView {
			case msgListView:
				m.msgDetailsVP.GotoTop()
			case msgDetailsView:
				if m.msgDetailsVP.AtTop() {
					break
				}
				m.msgDetailsVP.ScrollUp(viewPortMoveLineCount)
			case helpView:
				if m.helpVP.AtTop() {
					break
				}
				m.helpVP.ScrollUp(viewPortMoveLineCount)
			}
		case "tab", "shift+tab":
			if len(m.msgsList.Items()) == 0 {
				break
			}

			switch m.activeView {
			case msgListView:
				m.activeView = msgDetailsView
			case msgDetailsView:
				m.activeView = msgListView
			}
		case "P":
			if m.activeView == helpView {
				break
			}

			if len(m.msgsList.Items()) == 0 {
				m.errorMsg = "no item in list"
				break
			}

			message, ok := m.msgsList.SelectedItem().(t.Message)
			if !ok {
				m.errorMsg = genericErrMsg
				break
			}

			cmds = append(cmds, saveRecordDetailsToDisk(message, m.outputDir, m.config.Topic, true))
		}
	case tea.WindowSizeMsg:
		w1, h1 := messageListStyle.GetFrameSize()
		w2, h2 := viewPortStyle.GetFrameSize()
		m.msgsList.SetHeight(msg.Height - h1 - 3)

		fullScreenVPHeight := msg.Height - 6
		msgDetailsVPHeight := msg.Height - h2 - 2 - 3
		m.msgDetailsVPWidth = max(10, msg.Width-w1-w2-m.msgsList.Width())

		if !m.msgDetailsVPReady {
			m.msgDetailsVP = viewport.New(m.msgDetailsVPWidth, msgDetailsVPHeight)
			m.msgDetailsVP.KeyMap.HalfPageDown.SetKeys("ctrl+d")
			m.msgDetailsVP.KeyMap.Up.SetEnabled(false)
			m.msgDetailsVP.KeyMap.Down.SetEnabled(false)
			m.msgDetailsVPReady = true
		} else {
			m.msgDetailsVP.Width = m.msgDetailsVPWidth
			m.msgDetailsVP.Height = msgDetailsVPHeight
		}
		forceRefreshMsgDetailsVP = true

		helpVPWidth := msg.Width - w2 - 4
		if !m.helpVPReady {
			m.helpVP = viewport.New(helpVPWidth, fullScreenVPHeight)
			m.helpVP.SetContent(helpText)
			m.helpVP.KeyMap.HalfPageDown.SetKeys("ctrl+d")
			m.helpVP.KeyMap.Up.SetEnabled(false)
			m.helpVP.KeyMap.Down.SetEnabled(false)
			m.helpVPReady = true
		} else {
			m.helpVP.Width = helpVPWidth
			m.helpVP.Height = fullScreenVPHeight
		}

	case msgsFetchedMsg:
		m.fetchingInProgress = false
		if msg.err != nil {
			m.errorMsg = fmt.Sprintf("failed to fetch messages: %s", msg.err.Error())
			break
		}

		if len(msg.messages) == 0 {
			m.msg = "No new messages found"
			break
		}

		switch m.behaviours.SkipMessages {
		case false:
			for _, message := range msg.messages {
				m.msgsList.InsertItem(len(m.msgsList.Items()), message)
				if m.behaviours.PersistMessages {
					cmds = append(cmds, saveRecordDetailsToDisk(message, m.outputDir, m.config.Topic, false))
				}
			}
			m.msg = fmt.Sprintf("%d message(s) fetched", len(msg.messages))
		case true:
			m.msg = fmt.Sprintf("skipped over %d message(s)", len(msg.messages))
		}

	case msgSavedToDiskMsg:
		if msg.err != nil {
			m.errorMsg = fmt.Sprintf("Error saving to disk: %s", msg.err.Error())
		} else if msg.notifyUserOnSuccess {
			m.msg = "written to file"
		}
	case dataWrittenToClipboard:
		if msg.err != nil {
			m.errorMsg = fmt.Sprintf("couldn't copy details to clipboard: %s", msg.err)
		} else {
			m.msg = "details copied to clipboard!"
		}
	case hideHelpMsg:
		m.showHelpIndicator = false
	}

	switch m.activeView {
	case msgListView:
		m.msgsList, cmd = m.msgsList.Update(msg)
		cmds = append(cmds, cmd)
	case msgDetailsView:
		m.msgDetailsVP, cmd = m.msgDetailsVP.Update(msg)
		cmds = append(cmds, cmd)
	case helpView:
		m.helpVP, cmd = m.helpVP.Update(msg)
		cmds = append(cmds, cmd)
	}

	m.updateMsgDetailsVP(forceRefreshMsgDetailsVP)

	return m, tea.Batch(cmds...)
}

func (m *Model) updateMsgDetailsVP(forceRefresh bool) {
	if m.activeView == msgListView || m.activeView == msgDetailsView {
		if len(m.msgsList.Items()) > 0 && (forceRefresh || m.msgsList.Index() != m.currentMsgIndex) {
			m.currentMsgIndex = m.msgsList.Index()
			message, ok := m.msgsList.SelectedItem().(t.Message)

			if ok {
				m.msgDetailsVP.SetContent(getMsgDetailsStylized(message, m.config.Encoding, m.msgDetailsVPWidth))
			}

		}
	}
}
