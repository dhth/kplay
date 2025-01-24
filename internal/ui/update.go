package ui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

const (
	useHighPerformanceRenderer = false
	vpScrollLineChunk          = 3
)

const (
	msgAttributeNotFoundMsg = "something went wrong (with kplay)"
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
			switch m.vpFullScreen {
			case true:
				m.msgMetadataVP.Height = m.msgMetadataVPHeight
				m.msgValueVP.Height = m.msgValueVPHeight
				m.vpFullScreen = false
				m.activeView = m.lastView
				m.lastView = m.activeView
			case false:
				switch m.activeView {
				case kMsgsListView:
					return m, tea.Quit
				case kMsgMetadataView:
					m.activeView = kMsgsListView
				case kMsgValueView:
					m.activeView = kMsgMetadataView
				}
			}
		case "Q":
			return m, tea.Quit
		case "n", " ":
			cmds = append(cmds, FetchRecords(m.kCl, 1))
		case "N":
			cmds = append(cmds, FetchRecords(m.kCl, 10))
		case "}":
			cmds = append(cmds, FetchRecords(m.kCl, 100))
		case "?":
			m.lastView = m.activeView
			m.activeView = helpView
			m.vpFullScreen = true
		case "p":
			if !m.persistRecords {
				m.skipRecords = false
			}
			m.persistRecords = !m.persistRecords
		case "s":
			if !m.skipRecords {
				m.persistRecords = false
			}
			m.skipRecords = !m.skipRecords
		case "1", "m":
			m.msgMetadataVP.Height = m.terminalHeight - 7
			m.vpFullScreen = true
			m.lastView = m.activeView
			m.activeView = kMsgMetadataView
			m.msgMetadataVP.GotoTop()
		case "2", "v":
			m.msgValueVP.Height = m.terminalHeight - 7
			m.vpFullScreen = true
			m.lastView = m.activeView
			m.activeView = kMsgValueView
			m.msgValueVP.GotoTop()
		case "[":
			m.kMsgsList.CursorUp()
			m.msgMetadataVP.SetContent(m.recordMetadataStore[m.kMsgsList.SelectedItem().FilterValue()])
			m.msgValueVP.SetContent(m.recordValueStore[m.kMsgsList.SelectedItem().FilterValue()])
		case "]":
			m.kMsgsList.CursorDown()
			m.msgMetadataVP.SetContent(m.recordMetadataStore[m.kMsgsList.SelectedItem().FilterValue()])
			m.msgValueVP.SetContent(m.recordValueStore[m.kMsgsList.SelectedItem().FilterValue()])
		case "J":
			switch m.activeView {
			case kMsgsListView, kMsgValueView:
				m.msgValueVP.LineDown(vpScrollLineChunk)
			default:
				m.msgMetadataVP.LineDown(vpScrollLineChunk)
			}
		case "K":
			switch m.activeView {
			case kMsgsListView, kMsgValueView:
				m.msgValueVP.LineUp(vpScrollLineChunk)
			default:
				m.msgMetadataVP.LineUp(vpScrollLineChunk)
			}
		case "f":
			switch m.activeView {
			case kMsgMetadataView:
				switch m.vpFullScreen {
				case false:
					m.msgMetadataVP.Height = m.terminalHeight - 7
					m.lastView = kMsgMetadataView
					m.vpFullScreen = true
				case true:
					m.msgMetadataVP.Height = m.msgMetadataVPHeight
					m.msgValueVP.Height = m.msgValueVPHeight
					m.vpFullScreen = false
					m.activeView = m.lastView
				}
				m.msgMetadataVP.GotoTop()
			case kMsgValueView:
				switch m.vpFullScreen {
				case false:
					m.msgValueVP.Height = m.terminalHeight - 7
					m.lastView = kMsgValueView
					m.vpFullScreen = true
				case true:
					m.msgValueVP.Height = m.msgValueVPHeight
					m.msgMetadataVP.Height = m.msgMetadataVPHeight
					m.vpFullScreen = false
					m.activeView = m.lastView
				}
				m.msgValueVP.GotoTop()
			}
		case "tab":
			if m.vpFullScreen {
				break
			}
			if m.activeView == kMsgsListView {
				m.activeView = kMsgMetadataView
			} else if m.activeView == kMsgMetadataView {
				m.activeView = kMsgValueView
			} else if m.activeView == kMsgValueView {
				m.activeView = kMsgsListView
			}
		case "shift+tab":
			if m.vpFullScreen {
				break
			}
			if m.activeView == kMsgsListView {
				m.activeView = kMsgValueView
			} else if m.activeView == kMsgMetadataView {
				m.activeView = kMsgsListView
			} else if m.activeView == kMsgValueView {
				m.activeView = kMsgMetadataView
			}
		}

	case tea.WindowSizeMsg:
		w1, h1 := messageListStyle.GetFrameSize()
		w2, h2 := viewPortFullScreenStyle.GetFrameSize()
		m.terminalHeight = msg.Height
		m.terminalWidth = msg.Width
		m.kMsgsList.SetHeight(msg.Height - h1 - 2)

		fullScreenVPHeight := msg.Height - 7
		switch m.vpFullScreen {
		case true:
			m.msgMetadataVPHeight = fullScreenVPHeight
			m.msgValueVPHeight = fullScreenVPHeight
		case false:
			m.msgMetadataVPHeight = 6
			m.msgValueVPHeight = msg.Height - h2 - 2 - m.msgMetadataVPHeight - 8

		}
		vpWidth := msg.Width - w1 - w2 - 2

		if !m.msgMetadataVPReady {
			m.msgMetadataVP = viewport.New(vpWidth, m.msgMetadataVPHeight)
			m.msgMetadataVP.HighPerformanceRendering = useHighPerformanceRenderer
			m.msgMetadataVPReady = true
		} else {
			m.msgMetadataVP.Width = vpWidth
			m.msgMetadataVP.Height = m.msgMetadataVPHeight
		}

		if !m.msgValueVPReady {
			m.msgValueVP = viewport.New(vpWidth, m.msgValueVPHeight)
			m.msgValueVP.HighPerformanceRendering = useHighPerformanceRenderer
			m.msgValueVPReady = true
		} else {
			m.msgValueVP.Width = vpWidth
			m.msgValueVP.Height = m.msgValueVPHeight
		}

		if !m.helpVPReady {
			m.helpVP = viewport.New(vpWidth, fullScreenVPHeight)
			m.helpVP.HighPerformanceRendering = useHighPerformanceRenderer
			m.helpVP.SetContent(helpText)
			m.helpVPReady = true
		} else {
			m.helpVP.Width = vpWidth
			m.helpVP.Height = fullScreenVPHeight
		}
	case KMsgMetadataReadyMsg:
		m.recordMetadataStore[msg.storeKey] = msg.msgMetadata

		if !m.firstMsgMetadataSet {
			m.msgMetadataVP.SetContent(msg.msgMetadata)
			m.kMsgsCurrentIndex = m.kMsgsList.Index()
			m.firstMsgMetadataSet = true
		}
		if m.persistRecords {
			cmds = append(cmds, saveRecordMetadataToDisk(msg.record, msg.msgMetadata))
		}

	case KMsgValueReadyMsg:
		if msg.err != nil {
			m.errorMsg = msg.err.Error()
		} else {
			m.recordValueStore[msg.storeKey] = msg.msgValue
			if !m.firstMsgValueSet {
				m.msgValueVP.SetContent(msg.msgValue)
				m.kMsgsCurrentIndex = m.kMsgsList.Index()
				m.firstMsgValueSet = true
			}
			if m.persistRecords {
				cmds = append(cmds, saveRecordValueToDisk(msg.record, msg.msgValue))
			}
		}

	case KMsgFetchedMsg:
		if msg.err != nil {
			m.errorMsg = msg.err.Error()
		} else {
			if len(msg.records) == 0 {
				m.msg = "No new messages found"
			} else {
				switch m.skipRecords {
				case false:
					for _, rec := range msg.records {
						m.kMsgsList.InsertItem(len(m.kMsgsList.Items()), KMsgItem{record: *rec})
						cmds = append(cmds, saveRecordMetadata(rec), saveRecordValue(rec, m.config.DeserFmt))
					}
					m.msg = fmt.Sprintf("%d message(s) fetched", len(msg.records))
				case true:
					m.msg = fmt.Sprintf("skipped over %d message(s)", len(msg.records))
				}
			}
		}

	case RecordSavedToDiskMsg:
		if msg.err != nil {
			m.errorMsg = fmt.Sprintf("Error saving to disk: %s", msg.err.Error())
		}
	case HideHelpMsg:
		m.showHelpIndicator = false
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
	case helpView:
		m.helpVP, cmd = m.helpVP.Update(msg)
		cmds = append(cmds, cmd)
	}

	numItems := len(m.kMsgsList.Items())
	msgIndex := m.kMsgsList.Index()

	if numItems > 0 && msgIndex != m.kMsgsCurrentIndex {
		msgID := m.kMsgsList.SelectedItem().FilterValue()
		msgMetadata, metadataOk := m.recordMetadataStore[msgID]
		if metadataOk {
			m.msgMetadataVP.SetContent(msgMetadata)
		} else {
			m.msgMetadataVP.SetContent(msgAttributeNotFoundMsg)
		}

		msgValue, metadataOk := m.recordValueStore[msgID]
		if metadataOk {
			m.msgValueVP.SetContent(msgValue)
		} else {
			m.msgValueVP.SetContent(msgAttributeNotFoundMsg)
		}

		m.kMsgsCurrentIndex = msgIndex
	}

	return m, tea.Batch(cmds...)
}
