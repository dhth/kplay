package model

import (
	"fmt"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

const useHighPerformanceRenderer = false

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd
	m.msg = ""
	m.errorMsg = ""

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "n", " ":
			return m, FetchNextKMsg(m.kCl, 1)
		case "N":
			return m, FetchNextKMsg(m.kCl, 10)
		case "p":
			m.persistRecords = !m.persistRecords
			return m, nil
		case "}":
			return m, FetchNextKMsg(m.kCl, 2000)
		case "f":
			switch m.activeView {
			case kMsgMetadataView:
				switch m.vpFullScreen {
				case false:
					m.msgMetadataVP.Height = m.terminalHeight - 8
					m.vpFullScreen = true
				case true:
					m.msgMetadataVP.Height = 12
					m.vpFullScreen = false
				}
			case kMsgValueView:
				switch m.vpFullScreen {
				case false:
					m.msgValueVP.Height = m.terminalHeight - 8
					m.vpFullScreen = true
				case true:
					m.msgValueVP.Height = 12
					m.vpFullScreen = false
				}
				return m, tea.Batch(cmds...)
			}
		case "tab":
			if m.activeView == kMsgsListView {
				m.activeView = kMsgMetadataView
			} else if m.activeView == kMsgMetadataView {
				m.activeView = kMsgValueView
			} else if m.activeView == kMsgValueView {
				m.activeView = kMsgsListView
			}
		case "shift+tab":
			if m.activeView == kMsgsListView {
				m.activeView = kMsgValueView
			} else if m.activeView == kMsgMetadataView {
				m.activeView = kMsgsListView
			} else if m.activeView == kMsgValueView {
				m.activeView = kMsgMetadataView
			}
		}

	case tea.WindowSizeMsg:
		_, h := stackListStyle.GetFrameSize()
		m.terminalHeight = msg.Height
		m.terminalWidth = msg.Width
		m.kMsgsList.SetHeight(msg.Height - h - 2)

		if !m.msgMetadataVPReady {
			m.msgMetadataVP = viewport.New(120, m.terminalHeight/2-8)
			m.msgMetadataVP.HighPerformanceRendering = useHighPerformanceRenderer
			m.msgMetadataVPReady = true
		} else {
			m.msgMetadataVP.Width = 120
			m.msgMetadataVP.Height = 12
		}

		if !m.msgValueVPReady {
			m.msgValueVP = viewport.New(120, m.terminalHeight/2-8)
			m.msgValueVP.HighPerformanceRendering = useHighPerformanceRenderer
			m.msgValueVPReady = true
		} else {
			m.msgValueVP.Width = 120
			m.msgValueVP.Height = 12
		}

	case KMsgDataReadyMsg:
		if msg.err != nil {
			m.errorMsg = msg.err.Error()
		} else {
			m.recordMetadataStore[msg.storeKey] = msg.msgMetadata
			m.recordValueStore[msg.storeKey] = msg.msgValue
			if m.persistRecords {
				cmds = append(cmds, SaveRecordToDisk(msg.record, msg.msgMetadata, msg.msgValue))
			}
		}
		return m, tea.Batch(cmds...)

	case KMsgFetchedMsg:
		if msg.err != nil {
			m.errorMsg = msg.err.Error()
		} else {
			if len(msg.records) == 0 {
				m.msg = "No new messages found"
			} else {
				for _, rec := range msg.records {
					m.kMsgsList.InsertItem(len(m.kMsgsList.Items()), KMsgItem{record: *rec})
					cmds = append(cmds, saveRecordData(rec))
				}
				m.msg = fmt.Sprintf("%d message(s) fetched", len(msg.records))
			}
		}
	case KMsgChosenMsg:
		m.msgMetadataVP.SetContent(m.recordMetadataStore[msg.key])
		m.msgValueVP.SetContent(m.recordValueStore[msg.key])

	case RecordSavedToDiskMsg:
		if msg.err != nil {
			m.errorMsg = fmt.Sprintf("Error saving to disk: %s", msg.err.Error())
		}
	}

	switch m.activeView {
	case kMsgsListView:
		m.kMsgsList, cmd = m.kMsgsList.Update(msg)
		cmds = append(cmds, cmd)
	case kMsgMetadataView:
		m.msgMetadataVP, cmd = m.msgMetadataVP.Update(msg)
		cmds = append(cmds, cmd)
	case kMsgValueView:
		m.msgValueVP, cmd = m.msgValueVP.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}
