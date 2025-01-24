package ui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

const (
	useHighPerformanceRenderer = false
	viewPortMoveLineCount      = 3
	msgAttributeNotFoundMsg    = "something went wrong (with kplay)"
	genericErrMsg              = "something went wrong"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd
	m.msg = ""
	m.errorMsg = ""

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			switch m.activeView {
			case kMsgsListView:
				return m, tea.Quit
			case kMsgValueView:
				m.activeView = kMsgsListView
			case helpView:
				m.activeView = m.lastView
			}
		case "Q":
			return m, tea.Quit
		case "n", " ":
			if m.activeView == helpView {
				break
			}

			cmds = append(cmds, FetchRecords(m.client, 1))
		case "N":
			if m.activeView == helpView {
				break
			}

			cmds = append(cmds, FetchRecords(m.client, 10))
		case "}":
			if m.activeView == helpView {
				break
			}

			cmds = append(cmds, FetchRecords(m.client, 100))
		case "?":
			m.lastView = m.activeView
			m.activeView = helpView
		case "p":
			if m.activeView == helpView {
				break
			}

			if !m.persistRecords {
				m.skipRecords = false
			}
			m.persistRecords = !m.persistRecords
		case "s":
			if m.activeView == helpView {
				break
			}

			if !m.skipRecords {
				m.persistRecords = false
			}
			m.skipRecords = !m.skipRecords
		case "y":
			if len(m.msgsList.Items()) == 0 {
				break
			}

			if m.activeView == helpView {
				break
			}

			item := m.msgsList.SelectedItem()
			if item == nil {
				break
			}

			details, ok := m.msgDetailsStore[item.FilterValue()]
			if !ok {
				m.errorMsg = genericErrMsg
				break
			}

			detailsStr := getMsgDetails(details)
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

			item := m.msgsList.SelectedItem()
			if item == nil {
				break
			}

			details := m.msgDetailsStore[item.FilterValue()]
			detailsStr := getMsgDetailsStylized(details)
			m.msgDetailsVP.SetContent(detailsStr)
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
			item := m.msgsList.SelectedItem()
			if item == nil {
				break
			}

			details := m.msgDetailsStore[item.FilterValue()]
			detailsStr := getMsgDetailsStylized(details)
			m.msgDetailsVP.SetContent(detailsStr)
		case "j", "down":
			switch m.activeView {
			case kMsgValueView:
				if m.msgDetailsVP.AtBottom() {
					break
				}
				m.msgDetailsVP.LineDown(viewPortMoveLineCount)
			case helpView:
				if m.helpVP.AtBottom() {
					break
				}
				m.helpVP.LineDown(viewPortMoveLineCount)
			}
		case "k", "up":
			switch m.activeView {
			case kMsgValueView:
				if m.msgDetailsVP.AtTop() {
					break
				}
				m.msgDetailsVP.LineUp(viewPortMoveLineCount)
			case helpView:
				if m.helpVP.AtTop() {
					break
				}
				m.helpVP.LineUp(viewPortMoveLineCount)
			}
		case "tab", "shift+tab":
			if len(m.msgsList.Items()) == 0 {
				break
			}

			switch m.activeView {
			case kMsgsListView:
				m.activeView = kMsgValueView
			case kMsgValueView:
				m.activeView = kMsgsListView
			}
		}
	case tea.WindowSizeMsg:
		w1, h1 := messageListStyle.GetFrameSize()
		w2, h2 := viewPortStyle.GetFrameSize()
		m.terminalHeight = msg.Height
		m.terminalWidth = msg.Width
		m.msgsList.SetHeight(msg.Height - h1 - 3)

		fullScreenVPHeight := msg.Height - 6
		msgDetailsVPHeight := msg.Height - h2 - 2 - 3
		msgDetailsVPWidth := msg.Width - w1 - w2 - m.msgsList.Width() - 2

		if !m.msgDetailsVPReady {
			m.msgDetailsVP = viewport.New(msgDetailsVPWidth, msgDetailsVPHeight)
			m.msgDetailsVP.HighPerformanceRendering = useHighPerformanceRenderer
			m.msgDetailsVP.KeyMap.HalfPageDown.SetKeys("ctrl+d")
			m.msgDetailsVP.KeyMap.Up.SetEnabled(false)
			m.msgDetailsVP.KeyMap.Down.SetEnabled(false)
			m.msgDetailsVPReady = true
		} else {
			m.msgDetailsVP.Width = msgDetailsVPWidth
			m.msgDetailsVP.Height = msgDetailsVPHeight
		}

		helpVPWidth := msg.Width - w2 - 4
		if !m.helpVPReady {
			m.helpVP = viewport.New(helpVPWidth, fullScreenVPHeight)
			m.helpVP.HighPerformanceRendering = useHighPerformanceRenderer
			m.helpVP.SetContent(helpText)
			m.helpVP.KeyMap.HalfPageDown.SetKeys("ctrl+d")
			m.helpVP.KeyMap.Up.SetEnabled(false)
			m.helpVP.KeyMap.Down.SetEnabled(false)
			m.helpVPReady = true
		} else {
			m.helpVP.Width = helpVPWidth
			m.helpVP.Height = fullScreenVPHeight
		}

	case msgDataReadyMsg:
		m.msgDetailsStore[msg.uniqueKey] = msg.details

		if !m.firstMsgValueSet {
			firstItem := m.msgsList.SelectedItem()
			if firstItem != nil {
				msgDetails, ok := m.msgDetailsStore[firstItem.FilterValue()]
				if ok {
					detailsStylized := getMsgDetailsStylized(msgDetails)
					m.msgDetailsVP.SetContent(detailsStylized)
					m.currentMsgIndex = m.msgsList.Index()
					m.firstMsgValueSet = true
				}
			}
		}

		if m.persistRecords {
			details := getMsgDetails(msg.details)
			cmds = append(cmds, saveRecordValueToDisk(msg.uniqueKey, details))
		}

	case msgFetchedMsg:
		if msg.err != nil {
			m.errorMsg = msg.err.Error()
			break
		}

		if len(msg.records) == 0 {
			m.msg = "No new messages found"
			break
		}

		switch m.skipRecords {
		case false:
			for _, rec := range msg.records {
				m.msgsList.InsertItem(len(m.msgsList.Items()), KMsgItem{record: *rec})
				cmds = append(cmds, generateRecordDetails(rec, m.config.DeserFmt))
			}
			m.msg = fmt.Sprintf("%d message(s) fetched", len(msg.records))
		case true:
			m.msg = fmt.Sprintf("skipped over %d message(s)", len(msg.records))
		}

	case msgSavedToDiskMsg:
		if msg.err != nil {
			m.errorMsg = fmt.Sprintf("Error saving to disk: %s", msg.err.Error())
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
	case kMsgsListView:
		m.msgsList, cmd = m.msgsList.Update(msg)
		cmds = append(cmds, cmd)
	case kMsgValueView:
		m.msgDetailsVP, cmd = m.msgDetailsVP.Update(msg)
		cmds = append(cmds, cmd)
	case helpView:
		m.helpVP, cmd = m.helpVP.Update(msg)
		cmds = append(cmds, cmd)
	}

	numItems := len(m.msgsList.Items())
	msgIndex := m.msgsList.Index()

	if numItems > 0 && msgIndex != m.currentMsgIndex {
		msgID := m.msgsList.SelectedItem().FilterValue()
		msgDetails, ok := m.msgDetailsStore[msgID]
		if ok {
			detailsStylized := getMsgDetailsStylized(msgDetails)
			m.msgDetailsVP.SetContent(detailsStylized)
		} else {
			m.msgDetailsVP.SetContent(msgAttributeNotFoundMsg)
		}

		m.currentMsgIndex = msgIndex
	}

	return m, tea.Batch(cmds...)
}
