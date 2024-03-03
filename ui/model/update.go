package model

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/dhth/kplay/ui/model/generated"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
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
		case "}":
			return m, FetchNextKMsg(m.kCl, 100)
		case "ctrl+f":
			switch m.activeView {
			case kMsgHeaderView:
				switch m.vpFullScreen {
				case false:
					m.msgHeadersVP.Height = m.terminalHeight - 8
					m.vpFullScreen = true
				case true:
					m.msgHeadersVP.Height = 12
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
				m.activeView = kMsgHeaderView
			} else if m.activeView == kMsgHeaderView {
				m.activeView = kMsgValueView
			} else if m.activeView == kMsgValueView {
				m.activeView = kMsgsListView
			}
		case "shift+tab":
			if m.activeView == kMsgsListView {
				m.activeView = kMsgValueView
			} else if m.activeView == kMsgHeaderView {
				m.activeView = kMsgsListView
			} else if m.activeView == kMsgValueView {
				m.activeView = kMsgHeaderView
			}
		}

	case tea.WindowSizeMsg:
		_, h := stackListStyle.GetFrameSize()
		m.terminalHeight = msg.Height
		m.kMsgsList.SetHeight(msg.Height - h - 2)

		if !m.msgHeadersVPReady {
			m.msgHeadersVP = viewport.New(120, 12)
			m.msgHeadersVP.HighPerformanceRendering = useHighPerformanceRenderer
			m.msgHeadersVPReady = true
		} else {
			m.msgHeadersVP.Width = 120
			m.msgHeadersVP.Height = 12
		}

		if !m.msgValueVPReady {
			m.msgValueVP = viewport.New(120, 12)
			m.msgValueVP.HighPerformanceRendering = useHighPerformanceRenderer
			m.msgValueVPReady = true
		} else {
			m.msgValueVP.Width = 120
			m.msgValueVP.Height = 12
		}

	case KMsgChosenMsg:
		if len(msg.item.record.Value) == 0 {
			m.msgValueVP.SetContent("Tombstone")
		} else {
			message := &generated.ApplicationState{}
			if err := proto.Unmarshal(msg.item.record.Value, message); err != nil {
				m.errorMsg = fmt.Sprintf("Failed to deserialize message")
			} else {
				jsonData, err := protojson.Marshal(message)
				if err != nil {
					m.errorMsg = "Failed to marshal message to JSON"
				} else {
					headersJSON, err := json.MarshalIndent(msg.item.record.Headers, "", "    ")
					if err == nil {
						m.msgHeadersVP.SetContent(string(headersJSON))
					}

					var cont bytes.Buffer
					err = json.Indent(&cont, jsonData, "", "    ")
					c := cont.String()
					m.msgValueVP.SetContent(c)
				}
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
				}
				m.msg = fmt.Sprintf("%d message(s) fetched", len(msg.records))

			}
		}

	}

	switch m.activeView {
	case kMsgsListView:
		m.kMsgsList, cmd = m.kMsgsList.Update(msg)
		cmds = append(cmds, cmd)
	case kMsgHeaderView:
		m.msgHeadersVP, cmd = m.msgHeadersVP.Update(msg)
		cmds = append(cmds, cmd)
	case kMsgValueView:
		m.msgValueVP, cmd = m.msgValueVP.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}
